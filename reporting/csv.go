package reporting

import (
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"studio-console/domain"
)

type CSVOptions struct {
	IncludeID      bool
	IncludeContact bool
	IncludeStatus  bool
	Location       *time.Location
}

func DefaultCSVOptions() CSVOptions {
	return CSVOptions{IncludeID: true, IncludeContact: true, IncludeStatus: true, Location: time.UTC}
}

func WriteStaffCSV(writer io.Writer, accounts []domain.StaffAccount, options CSVOptions) error {
	if writer == nil {
		return errors.New("CSV writer is required")
	}
	if options.Location == nil {
		options.Location = time.UTC
	}
	rows := BuildStaffRows(accounts, options.Location)
	headings := make([]string, 0, 9)
	if options.IncludeID {
		headings = append(headings, "人员编号")
	}
	headings = append(headings, "姓名", "角色")
	if options.IncludeContact {
		headings = append(headings, "手机", "邮箱")
	}
	if options.IncludeStatus {
		headings = append(headings, "状态")
	}
	headings = append(headings, "创建时间", "更新时间")
	csvWriter := csv.NewWriter(writer)
	if err := csvWriter.Write(headings); err != nil {
		return err
	}
	for _, row := range rows {
		values := make([]string, 0, len(headings))
		if options.IncludeID {
			values = append(values, row.ID)
		}
		values = append(values, row.Name, row.RoleDisplay)
		if options.IncludeContact {
			values = append(values, row.Phone, row.Email)
		}
		if options.IncludeStatus {
			values = append(values, row.StatusLabel)
		}
		values = append(values, row.CreatedAt, row.UpdatedAt)
		if err := csvWriter.Write(values); err != nil {
			return err
		}
	}
	csvWriter.Flush()
	return csvWriter.Error()
}

type ImportRow struct {
	Line   int
	Name   string
	Phone  string
	Email  string
	Role   string
	Status string
}

type ImportProblem struct {
	Line    int    `json:"line"`
	Message string `json:"message"`
}

func ReadStaffCSV(reader io.Reader, maxRows int) ([]ImportRow, []ImportProblem, error) {
	if reader == nil {
		return nil, nil, errors.New("CSV reader is required")
	}
	if maxRows < 1 {
		return nil, nil, errors.New("maximum row count must be positive")
	}
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(records) == 0 {
		return nil, nil, errors.New("CSV file is empty")
	}
	headings := indexHeadings(records[0])
	required := []string{"姓名", "角色"}
	for _, name := range required {
		if _, exists := headings[name]; !exists {
			return nil, nil, errors.New("CSV 缺少列: " + name)
		}
	}
	rows := make([]ImportRow, 0)
	problems := make([]ImportProblem, 0)
	for index, record := range records[1:] {
		line := index + 2
		if len(rows) >= maxRows {
			problems = append(problems, ImportProblem{Line: line, Message: "超过最大导入行数"})
			break
		}
		row := ImportRow{Line: line, Name: cell(record, headings, "姓名"), Phone: cell(record, headings, "手机"), Email: cell(record, headings, "邮箱"), Role: cell(record, headings, "角色"), Status: cell(record, headings, "状态")}
		if strings.TrimSpace(row.Name) == "" {
			problems = append(problems, ImportProblem{Line: line, Message: "姓名不能为空"})
			continue
		}
		if strings.TrimSpace(row.Role) == "" {
			problems = append(problems, ImportProblem{Line: line, Message: "角色不能为空"})
			continue
		}
		rows = append(rows, row)
	}
	return rows, problems, nil
}

func indexHeadings(record []string) map[string]int {
	result := make(map[string]int)
	for index, value := range record {
		result[strings.TrimSpace(value)] = index
	}
	return result
}

func cell(record []string, headings map[string]int, name string) string {
	index, exists := headings[name]
	if !exists || index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func FormatImportSummary(imported int, problems []ImportProblem) string {
	return "成功导入 " + strconv.Itoa(imported) + " 条，失败 " + strconv.Itoa(len(problems)) + " 条"
}
