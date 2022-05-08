package report

import (
	"cube/gologger"
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"strconv"
	"strings"
)

func convertXY(x, y int) string {
	return fmt.Sprintf(string(rune(x)) + strconv.Itoa(y))
}

func GetCsvShellValue(ip, module string, csvs []CsvCell) (s string) {
	//传入module 和 ip获取到string
	for _, csv := range csvs {
		if csv.Ip == ip && csv.Module == module {
			s = csv.Cell
			break
		} else {
			s = ""
		}
	}
	return s
}

func RemoveDuplicateCSS(css []CsvCell) []CsvCell {
	//struct slice去重, 利用map的key不能重复的特性
	resultMap := map[string]bool{}
	for _, v := range css {
		data, _ := json.Marshal(v)
		resultMap[string(data)] = true
	}
	var result []CsvCell
	for k := range resultMap {
		var t CsvCell
		json.Unmarshal([]byte(k), &t)
		result = append(result, t)
	}
	return result
}

func RemoveDuplicateResult(old []CsvCell) []CsvCell {
	var result []CsvCell
	for i := range old {
		flag := true
		for j := range result {
			if old[i].Ip == result[j].Ip && old[i].Module == result[j].Module {
				flag = false
				break
			}
		}
		if flag {
			result = append(result, old[i])
		}
	}
	return result
}

func WriteExportExcel(ccs []CsvCell, fp string) {
	plugMap := SortPlug(ccs)
	ipMap := SortIP(ccs)

	heads := GetKeys(plugMap)
	ips := GetKeys(ipMap)

	//gologger.Debug("IP: ", ips)
	heads = append([]string{"IP"}, heads...)

	excel := excelize.NewFile()

	style, err := excel.NewStyle(`{
					"border":[
                                {
                                        "type":"left",
                                        "color":"000000",
                                        "style":1
                                },
                                {
                                        "type":"top",
                                        "color":"000000",
                                        "style":1
                                },
                                {
                                        "type":"bottom",
                                        "color":"000000",
                                        "style":1
                                },
                                {
                                        "type":"right",
                                        "color":"000000",
                                        "style":1
                                }
                        ],
            "alignment":{
                "wrap_text":true
            }
        }`)

	styleIP, err := excel.NewStyle(`{
					"border":[
                                {
                                        "type":"left",
                                        "color":"000000",
                                        "style":1
                                },
                                {
                                        "type":"top",
                                        "color":"000000",
                                        "style":1
                                },
                                {
                                        "type":"bottom",
                                        "color":"000000",
                                        "style":1
                                },
                                {
                                        "type":"right",
                                        "color":"000000",
                                        "style":1
                                }
                        ],
            "alignment":{
				"horizontal":"center",
				"vertical":"center",
 				"wrap_text":true
            }
        }`)
	_ = excel.SetSheetRow("Sheet1", "A1", &heads)
	//_ = excel.SetRowStyle("Sheet1", 0, len(heads), styleIP)
	if err != nil {
		gologger.Error(err)
	}

	_ = excel.SetColWidth("Sheet1", "A", "A", 15)
	y := 2
	for _, ip := range ips {
		x := 65
		_ = excel.SetCellStyle("Sheet1", fmt.Sprintf("A%d", y), fmt.Sprintf("A%d", y), styleIP)
		_ = excel.SetCellValue("Sheet1", fmt.Sprintf("A%d", y), ip)
		x += 1
		for _, plug := range heads[1:] {
			data := GetCsvShellValue(ip, plug, ccs)
			if len(data) > 0 {
				_ = excel.SetColWidth("Sheet1", string(rune(x)), string(rune(x)), 30)
				_ = excel.SetCellStyle("Sheet1", convertXY(x, y), convertXY(x, y), style)
				_ = excel.SetCellValue("Sheet1", convertXY(x, y), strings.Trim(data, " "))
				x += 1
			} else {
				_ = excel.SetCellStyle("Sheet1", convertXY(x, y), convertXY(x, y), style)
				x += 1
			}
		}
		y += 1
	}

	if err := excel.SaveAs(fp); err != nil {
		gologger.Errorf("write to %s error: %s", fp, err)
	}
	excel.Close()
}

func ReadExportExcel(fp string) (ccs []CsvCell) {
	var headers []string
	var ip string

	excel, err := excelize.OpenFile(fp)
	if err != nil {
		gologger.Errorf("read %s error: %s", fp, err)
		return
	}
	rows, _ := excel.GetRows("Sheet1")
	for k, row := range rows {
		//fmt.Println(row)
		if k == 0 {
			for _, colCell := range row {
				headers = append(headers, colCell)
			}
		} else {
			for key, colCell := range row {
				if key == 0 {
					ip = colCell
				} else {
					cc := CsvCell{
						Ip:     ip,
						Module: headers[key],
						Cell:   colCell,
					}
					ccs = append(ccs, cc)
				}
				//fmt.Printf("Key: %d Col: %s\n", key, colCell)
			}
		}

	}
	return ccs
}
