package svc

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"sort"

	"github.com/xuri/excelize/v2"
)

var VarNumberMap = map[string]byte{
	"PLU":         0,
	"ProductName": 1,
	"Unit":        2,
	"TaxModel":    3,
	"TaxType":     4,
	"Price":       5,
	"UnitWeight":  6,
	"PreTare":     7,
	"LimitHigh":   8,
	"LimitLow":    9,
	"isUSED":      255,
	// "ProductCode": 10,
	// "ItemCode":    11,
}

var VarLenthMap = map[string]byte{
	"PLU":         4,
	"ProductName": 80,
	"Unit":        1,
	"TaxModel":    1,
	"TaxType":     1,
	"Price":       8,
	"UnitWeight":  8,
	"PreTare":     8,
	"LimitHigh":   8,
	"LimitLow":    8,
	"isUSED":      1,
	// "ProductCode": 4,
	// "ItemCode":    4,
}

type Product struct {
	PLU         int
	ProductName string
	Unit        byte
	TaxModel    byte
	TaxType     byte
	Price       float64
	UnitWeight  float64
	PreTare     float64
	LimitHigh   float64
	LimitLow    float64
	isUSED      byte
	ProductCode int
	ItemCode    int
}

/*
uint16_t  plu_serial;			0	4	PLU编号
uint8_t plu_name[PLU_NAME_MAX];	1	30	品名
uint8_t price_unit; 				2	1	价格单位（kg/100g/amount）
uint8_t tax_mode;				3	1	税的模式（外含/内含/无）
uint8_t tax_type;				4	1	税的类型（type1/type2/type3）
double unit_price; 				5	8	单价
double unit_weight;			6	8	单重
double ptare; 				7	8	预扣重
double plu_limit_high;			8	8	上限
double plu_limit_low;  			9	8	下限
*/

func ParserPluFile(excelFileName string, nameMaxLen int) bool {

	// columnNames := []string{
	// 	"PLU",
	// 	"ProductName",
	// 	"PriceUnit",
	// 	"TaxModel",
	// 	"TaxType",
	// 	"Price",
	// 	"UnitWeight",
	// 	"PreTare",
	// 	"LimitHigh",
	// 	"LimitLow",
	// 	"isUSED",
	// }

	md5Str := fileToMd5(excelFileName)

	products, err := readExcelToJSON(excelFileName)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	jsonData, err := json.Marshal(products)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	fmt.Println(string(jsonData))

	// 获取VarNumberMap长度
	mapLength := byte(len(VarNumberMap))
	// 创建bytes.Buffer
	buf := new(bytes.Buffer)

	buf.Write([]byte(md5Str))

	pluNameMaxLenth := nameMaxLen
	var tempProduct []Product
	//删除空行
	for _, product := range products {
		if product.PLU != 0 {
			tempProduct = append(tempProduct, product)
		}
	}
	sort.Slice(tempProduct, func(i, j int) bool {
		return tempProduct[i].PLU < tempProduct[j].PLU
	})
	products = tempProduct
	// 将plu数量写入bytes.Buffer
	pluCount := make([]byte, 2)
	binary.LittleEndian.PutUint16(pluCount, uint16(len(products)))
	buf.Write(pluCount)
	// 将长度写入bytes.Buffer
	buf.WriteByte(mapLength)

	varKeys := make([]string, 0, len(VarNumberMap))
	for key := range VarNumberMap {
		varKeys = append(varKeys, key)
	}
	sort.Slice(varKeys, func(i, j int) bool {
		return VarNumberMap[varKeys[i]] < VarNumberMap[varKeys[j]]
	})

	for _, key := range varKeys {
		// 按1个字节写入VarNumberMap的值
		buf.WriteByte(VarNumberMap[key])
		// 按2个字节写入VarLenthMap的值，以小端序列化
		b := make([]byte, 2)
		if key == "ProductName" {
			binary.LittleEndian.PutUint16(b, uint16(pluNameMaxLenth))
		} else {
			binary.LittleEndian.PutUint16(b, uint16(VarLenthMap[key]))
		}
		buf.Write(b)
	}

	for _, product := range products {
		// 按照VarLenthMap的字节数写入buf，以小端序列化
		// 写入ProductNumber，4字节
		binary.Write(buf, binary.LittleEndian, int32(product.PLU))
		// 写入ProductName，根据pluNameMaxLenth的值计算字节数
		if len(product.ProductName) > pluNameMaxLenth {
			var tempName = product.ProductName[:pluNameMaxLenth]
			productNameBytes := []byte(tempName)
			buf.Write(productNameBytes)
		} else {
			productNameBytes := []byte(product.ProductName)
			productNameBytes = append(productNameBytes, make([]byte, pluNameMaxLenth-len(productNameBytes))...)
			buf.Write(productNameBytes)
		}

		// 写入PriceUnit，1字节
		buf.WriteByte(byte(product.Unit))
		// 写入TaxModel，1字节
		buf.WriteByte(byte(product.TaxModel))
		// 写入TaxType，1字节
		buf.WriteByte(byte(product.TaxType))
		// 写入Price，8字节
		binary.Write(buf, binary.LittleEndian, float64(product.Price))
		// 写入UnitWeight，8字节
		binary.Write(buf, binary.LittleEndian, float64(product.UnitWeight))
		// 写入PreTare，8字节
		binary.Write(buf, binary.LittleEndian, float64(product.PreTare))
		// 写入LimitHigh，8字节
		binary.Write(buf, binary.LittleEndian, float64(product.LimitHigh))
		// 写入LimitLow，8字节
		binary.Write(buf, binary.LittleEndian, float64(product.LimitLow))

		// // 写入ProductCode ，4字节
		// binary.Write(buf, binary.LittleEndian, int32(product.ProductCode))
		// // 写入ItemCode ，4字节
		// binary.Write(buf, binary.LittleEndian, int32(product.ItemCode))
		// 写入isUSED，1字节  0xFF plu有效   0x00 plu无效
		buf.WriteByte(0xFF)
	}

	err = os.WriteFile("plu.bin", buf.Bytes(), 0644)
	if err != nil {
		fmt.Println("Write file error:", err)
		return false
	}
	print(products)
	print(pluNameMaxLenth)
	return true
}

func fileToMd5(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func readExcelToJSON(fileName string) ([]Product, error) {
	xlFile, err := excelize.OpenFile(fileName)
	if err != nil {
		return nil, err
	}

	var products []Product
	sheet1 := xlFile.GetSheetName(0)
	rows, err := xlFile.GetRows(sheet1)
	if err != nil {
		return nil, err
	}
	var colsList = rows[0]

	for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
		product := Product{}
		for i, colName := range colsList {
			cellValue, err := xlFile.GetCellValue(sheet1, indexToCellName(rowIndex+1, i+1))
			if err != nil {
				return products, err
			}
			switch colName {
			case "PLU":
				if cellValue != "" {
					product.PLU, err = convertToInt(cellValue)
					if err != nil {
						return products, err
					}
				} else {
					return products, fmt.Errorf("data error")
				}
			case "ProductCode":
				if cellValue != "" {
					product.ProductCode, err = convertToInt(cellValue)
					if err != nil {
						return products, err
					}
				} else {
					return products, fmt.Errorf("data error")
				}
			case "ItemCode":
				if cellValue != "" {
					product.ItemCode, err = convertToInt(cellValue)
					if err != nil {
						return products, err
					}
				} else {
					return products, fmt.Errorf("data error")
				}

			case "ProductName":
				if cellValue != "" {
					product.ProductName = cellValue
				} else {
					return products, fmt.Errorf("data error")
				}

			case "GeneralUnit":
				product.Unit, err = convertToByte(cellValue)
				if err != nil {
					return products, err
				}

			case "TaxModel":
				product.TaxModel, err = convertToByte(cellValue)
				if err != nil {
					return products, err
				}
			case "TaxType":
				product.TaxType, err = convertToByte(cellValue)
				if err != nil {
					return products, err
				}

			case "Price":
				product.Price, err = convertToFloat64(cellValue)
				if err != nil {
					return products, err
				}

			case "UnitWeight":
				product.UnitWeight, err = convertToFloat64(cellValue)
				if err != nil {
					return products, err
				}
			case "PreTare":
				product.PreTare, err = convertToFloat64(cellValue)
				if err != nil {
					return products, err
				}

			case "LimitHigh":
				product.LimitHigh, err = convertToFloat64(cellValue)
				if err != nil {
					return products, err
				}

			case "LimitLow":
				product.LimitLow, err = convertToFloat64(cellValue)
				if err != nil {
					return products, err
				}

			case "isUSED":
				product.isUSED, err = convertToByte(cellValue)
				if err != nil {
					return products, err
				}

			default:

			}
		}
		products = append(products, product)
	}
	return products, nil

}
func indexToCellName(row, col int) string {
	colName := ""
	for col > 0 {
		remainder := (col - 1) % 26
		colName = string(rune('A'+remainder)) + colName
		col = (col - 1) / 26
	}
	return colName + fmt.Sprint(row)
}

func convertToInt(value string) (int, error) {
	var result int
	if value != "" {
		_, err := fmt.Sscanf(value, "%d", &result)
		if err != nil {

			return result, err
		}
	}
	return result, nil
}

func convertToByte(value string) (byte, error) {
	var result byte
	if value != "" {
		_, err := fmt.Sscanf(value, "%d", &result)
		if err != nil {

			return result, err
		}
	}
	return result, nil
}

func convertToFloat64(value string) (float64, error) {
	var result float64
	if value != "" {
		_, err := fmt.Sscanf(value, "%f", &result)
		if err != nil {
			return result, err
		}
		result = math.Round(result*1000) / 1000
	}
	return result, nil
}

func ParserDelPlu(pluIdList []string) ([]byte, bool) {
	buf := new(bytes.Buffer)
	for i := 0; i < len(pluIdList); i++ {
		idNum, err := convertToInt(pluIdList[i])
		if err != nil {
			return buf.Bytes(), false
		}
		binary.Write(buf, binary.BigEndian, int32(idNum))
	}

	return buf.Bytes(), true
}

func ParserInsertPlu(excelFileName string, nameMaxLen int) ([]byte, []byte, bool) {
	// columnNames := []string{
	// 	"ProductNumber",
	// 	"ProductName",
	// 	"PriceUnit",
	// 	"TaxModel",
	// 	"TaxType",
	// 	"Price",
	// 	"UnitWeight",
	// 	"PreTare",
	// 	"LimitHigh",
	// 	"LimitLow",
	// 	"isUSED",
	// }

	pluNumBuf := new(bytes.Buffer)
	insertHeadBuf := new(bytes.Buffer)

	products, err := readExcelToJSON(excelFileName)
	if err != nil {
		fmt.Println("Error:", err)
		return pluNumBuf.Bytes(), insertHeadBuf.Bytes(), false
	}

	jsonData, err := json.Marshal(products)
	if err != nil {
		fmt.Println("Error:", err)
		return pluNumBuf.Bytes(), insertHeadBuf.Bytes(), false
	}

	fmt.Println(string(jsonData))
	// 获取VarNumberMap长度
	// mapLength := byte(len(VarNumberMap))
	// 创建bytes.Buffer
	buf := new(bytes.Buffer)
	// 将长度写入bytes.Buffer
	// buf.WriteByte(mapLength)
	pluNameMaxLenth := nameMaxLen
	var tempProduct []Product
	//删除空行
	for _, product := range products {
		if product.PLU != 0 {
			tempProduct = append(tempProduct, product)
		}
	}
	products = tempProduct

	varKeys := make([]string, 0, len(VarNumberMap))
	for key := range VarNumberMap {
		varKeys = append(varKeys, key)
	}
	sort.Slice(varKeys, func(i, j int) bool {
		return VarNumberMap[varKeys[i]] < VarNumberMap[varKeys[j]]
	})

	for _, key := range varKeys {
		// 按1个字节写入VarNumberMap的值
		// buf.WriteByte(VarNumberMap[key])
		insertHeadBuf.WriteByte(VarNumberMap[key])
		// 按2个字节写入VarLenthMap的值，以小端序列化
		b := make([]byte, 2)
		if key == "ProductName" {
			binary.LittleEndian.PutUint16(b, uint16(pluNameMaxLenth))
		} else {
			binary.LittleEndian.PutUint16(b, uint16(VarLenthMap[key]))
		}
		// buf.Write(b)
		insertHeadBuf.Write(b)
	}

	for _, product := range products {
		// 按照VarLenthMap的字节数写入buf，以小端序列化
		// 写入ProductNumber，4字节
		binary.Write(pluNumBuf, binary.BigEndian, int32(product.PLU)) //获取PLU number先删除，后插入
		binary.Write(buf, binary.LittleEndian, int32(product.PLU))
		// 写入ProductName，根据pluNameMaxLenth的值计算字节数
		if len(product.ProductName) > pluNameMaxLenth {
			var tempName = product.ProductName[:pluNameMaxLenth]
			productNameBytes := []byte(tempName)
			buf.Write(productNameBytes)
		} else {
			productNameBytes := []byte(product.ProductName)
			productNameBytes = append(productNameBytes, make([]byte, pluNameMaxLenth-len(productNameBytes))...)
			buf.Write(productNameBytes)
		}
		// 写入PriceUnit，1字节
		buf.WriteByte(byte(product.Unit))
		// 写入TaxModel，1字节
		buf.WriteByte(byte(product.TaxModel))
		// 写入TaxType，1字节
		buf.WriteByte(byte(product.TaxType))
		// 写入Price，8字节
		binary.Write(buf, binary.LittleEndian, float64(product.Price))
		// 写入UnitWeight，8字节
		binary.Write(buf, binary.LittleEndian, float64(product.UnitWeight))
		// 写入PreTare，8字节
		binary.Write(buf, binary.LittleEndian, float64(product.PreTare))
		// 写入LimitHigh，8字节
		binary.Write(buf, binary.LittleEndian, float64(product.LimitHigh))
		// 写入LimitLow，8字节
		binary.Write(buf, binary.LittleEndian, float64(product.LimitLow))
		// 写入isUSED，1字节  0xFF plu有效   0x00 plu无效
		buf.WriteByte(0xFF)
	}

	err = os.WriteFile("plu.bin", buf.Bytes(), 0644)
	if err != nil {
		fmt.Println("Write file error:", err)
		return pluNumBuf.Bytes(), insertHeadBuf.Bytes(), false
	}
	print(products)
	print(pluNameMaxLenth)
	return pluNumBuf.Bytes(), insertHeadBuf.Bytes(), true
}
