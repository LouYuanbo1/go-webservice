package multipart

import (
	"errors"
	"fmt"
	"mime/multipart"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/form/v4"
)

// ===== 正则表达式 =====
var (
	// 用于 snake_case 转换
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")

	// 索引解析正则
	indexedBracketPattern   = regexp.MustCompile(`^(.+)\[(\d+)\]\.(.+)$`) // field[0].sub
	indexedDotPattern       = regexp.MustCompile(`^(.+)\.(\d+)\.(.+)$`)   // field.0.sub
	indexOnlyBracketPattern = regexp.MustCompile(`^(.+)\[(\d+)\]$`)       // field[0]
	indexOnlyDotPattern     = regexp.MustCompile(`^(.+)\.(\d+)$`)         // field.0
)

// toSnakeCase 将 CamelCase 转为 snake_case（与 go-playground/form v4 默认转换一致）
func toSnakeCase(s string) string {
	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// extractFieldName 从 form 标签提取主字段名；无标签时转换为 snake_case
func extractFieldName(tag string, fieldName string) string {
	if tag != "" {
		if commaIdx := strings.Index(tag, ","); commaIdx != -1 {
			tag = tag[:commaIdx]
		}
		return tag
	}
	return toSnakeCase(fieldName)
}

// snakeCaseNameTransformer 是用于 form.Decoder 的名称转换函数，与 extractFieldName 逻辑一致
func snakeCaseNameTransformer(field reflect.StructField) string {
	tag := field.Tag.Get("form")
	return extractFieldName(tag, field.Name)
}

// BindMultipart 解析 multipart/form-data 请求，将文本字段和文件字段绑定到结构体 obj 中。
// obj 必须是可寻址的非 nil 指针。
//
// 重要约定（必须遵守）：
//  1. 嵌套字段路径分隔符必须使用 '.'（如 user.avatar），与 go-playground/form 行为一致
//     - 前端上传文件字段名需匹配此规则（如 axios 默认使用 '.'）
//     - 索引格式支持 bracket（photos[0]）或 dot（photos.0），但同一字段下禁止混用
//  2. 强烈建议为含文件的字段显式指定 form 标签（如 `form:"avatar"`）
//  3. 临时文件清理由 Gin 框架负责（请求结束时自动删除）
//  4. 匿名结构体字段提升规则严格遵循 Go 嵌入规范：
//     - 只有类型为结构体 T 或指向结构体的指针 *T 时，字段才提升（路径不拼接字段名）
//     - 其他类型（包括文件类型、基本类型等）均需拼接字段名
func BindMultipart[T any](gctx *gin.Context, obj T) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return errors.New("obj must be a non-nil pointer")
	}

	multipartForm, err := gctx.MultipartForm()
	if err != nil {
		return fmt.Errorf("failed to parse multipart form: %w", err)
	}

	// 创建解码器并注册统一的名称转换函数，确保文本字段与文件字段键名规则一致
	decoder := form.NewDecoder()
	decoder.RegisterTagNameFunc(snakeCaseNameTransformer)

	if err := decoder.Decode(obj, multipartForm.Value); err != nil {
		return fmt.Errorf("failed to decode form values: %w", err)
	}

	if err := fillFiles(obj, multipartForm.File); err != nil {
		return fmt.Errorf("failed to fill files: %w", err)
	}
	return nil
}

func fillFiles[T any](obj T, files map[string][]*multipart.FileHeader) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return errors.New("obj must be a non-nil pointer")
	}
	return fillFilesRecursive(v, v.Type(), files, "")
}

// fillFilesRecursive 核心递归函数
func fillFilesRecursive(v reflect.Value, t reflect.Type, files map[string][]*multipart.FileHeader, currentPath string) error {
	// ===== 情况0: 优先匹配文件类型（必须在指针解引用前判断）=====
	if v.Kind() == reflect.Slice && v.Type().Elem() == reflect.TypeFor[*multipart.FileHeader]() {
		return fillFileHeaderSlice(v, files, currentPath)
	}
	if v.Type() == reflect.TypeFor[*multipart.FileHeader]() {
		return fillSingleFileHeader(v, files, currentPath)
	}

	// ===== 情况1: 解引用指针 =====
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			// 仅当存在对应的文件键时才创建新值，避免将 nil 指针变为非 nil 零值
			if !hasFileKeyForPath(files, currentPath) {
				return nil
			}
			elemType := t.Elem()
			// 只为结构体或切片创建新值（文件类型已在情况0处理，其他类型如基本类型不会创建）
			if elemType.Kind() == reflect.Struct || elemType.Kind() == reflect.Slice {
				newVal := reflect.New(elemType)
				v.Set(newVal)
			} else {
				// 对于其他类型（如基本类型指针），保持 nil
				return nil
			}
		}
		return fillFilesRecursive(v.Elem(), t.Elem(), files, currentPath)
	}

	// ===== 情况2: 结构体 =====
	if v.Kind() == reflect.Struct {
		return fillStruct(v, t, files, currentPath)
	}

	// ===== 情况3: 结构体/指针结构体切片 =====
	if v.Kind() == reflect.Slice {
		elemType := v.Type().Elem()
		isStruct := elemType.Kind() == reflect.Struct
		isPtrToStruct := elemType.Kind() == reflect.Pointer && elemType.Elem().Kind() == reflect.Struct
		if isStruct || isPtrToStruct {
			return fillSlice(v, files, currentPath)
		}
		return nil // 其他类型切片跳过
	}

	return nil
}

// hasFileKeyForPath 检查 files 中是否存在任何需要当前路径的键（包括直接匹配或作为前缀）
func hasFileKeyForPath(files map[string][]*multipart.FileHeader, path string) bool {
	if path == "" {
		// 空路径仅在顶层递归时出现，但顶层 obj 已非 nil，此处不会触发；为安全返回 true
		return true
	}
	prefixDot := path + "."
	prefixBracket := path + "["
	for key := range files {
		if key == path || strings.HasPrefix(key, prefixDot) || strings.HasPrefix(key, prefixBracket) {
			return true
		}
	}
	return false
}

// fillFileHeaderSlice 处理 []*multipart.FileHeader 类型字段
func fillFileHeaderSlice(v reflect.Value, files map[string][]*multipart.FileHeader, currentPath string) error {
	noIndexFhs, hasNoIndex := files[currentPath]

	// 过滤空切片：如果无索引键存在但切片为空，视为不存在
	if hasNoIndex && len(noIndexFhs) == 0 {
		hasNoIndex = false
	}

	bracketMap := make(map[int][]*multipart.FileHeader)
	dotMap := make(map[int][]*multipart.FileHeader)

	for key, fhs := range files {
		// 仅当文件切片非空时才记录索引，避免空键导致误判冲突
		if len(fhs) == 0 {
			continue
		}
		if matches := indexOnlyBracketPattern.FindStringSubmatch(key); matches != nil && matches[1] == currentPath {
			if idx, err := strconv.Atoi(matches[2]); err == nil && idx >= 0 {
				if _, exists := bracketMap[idx]; exists {
					return fmt.Errorf("duplicate bracket index %d for path %s (key: %s)", idx, currentPath, key)
				}
				bracketMap[idx] = fhs
			}
		}
		if matches := indexOnlyDotPattern.FindStringSubmatch(key); matches != nil && matches[1] == currentPath {
			if idx, err := strconv.Atoi(matches[2]); err == nil && idx >= 0 {
				if _, exists := dotMap[idx]; exists {
					return fmt.Errorf("duplicate dot index %d for path %s (key: %s)", idx, currentPath, key)
				}
				dotMap[idx] = fhs
			}
		}
	}

	hasIndex := len(bracketMap) > 0 || len(dotMap) > 0
	if hasNoIndex && hasIndex {
		return fmt.Errorf("cannot provide both non-indexed key (%q) and indexed keys for path %s", currentPath, currentPath)
	}

	// 优先无索引键
	if hasNoIndex {
		newSlice := reflect.MakeSlice(v.Type(), len(noIndexFhs), len(noIndexFhs))
		for i, fh := range noIndexFhs {
			if fh != nil {
				newSlice.Index(i).Set(reflect.ValueOf(fh))
			}
		}
		v.Set(newSlice)
		return nil
	}

	if len(bracketMap) > 0 && len(dotMap) > 0 {
		return fmt.Errorf("cannot mix bracket and dot index formats for path %s", currentPath)
	}

	useMap := bracketMap
	if len(bracketMap) == 0 {
		useMap = dotMap
	}
	if len(useMap) == 0 {
		return nil
	}

	// 计算所需最小长度
	maxIdx := -1
	for idx := range useMap {
		if idx > maxIdx {
			maxIdx = idx
		}
	}
	needLen := maxIdx + 1

	// 如果当前切片长度不足，则扩展；否则直接使用原切片（保留尾部多余元素）
	if v.Len() < needLen {
		newSlice := reflect.MakeSlice(v.Type(), needLen, needLen)
		// 复制原有元素（仅复制到原长度，新位置为零值）
		for i := 0; i < v.Len(); i++ {
			newSlice.Index(i).Set(v.Index(i))
		}
		v.Set(newSlice)
	}

	// 填充索引对应的文件（此时 v.Len() >= needLen）
	for idx, fhs := range useMap {
		if len(fhs) == 0 {
			continue
		}
		if len(fhs) > 1 {
			return fmt.Errorf("multiple files uploaded for single index %d in path %s", idx, currentPath)
		}
		if idx < v.Len() && fhs[0] != nil {
			v.Index(idx).Set(reflect.ValueOf(fhs[0]))
		}
	}
	return nil
}

// fillSingleFileHeader 处理 *multipart.FileHeader 类型字段
func fillSingleFileHeader(v reflect.Value, files map[string][]*multipart.FileHeader, currentPath string) error {
	fhs, ok := files[currentPath]
	if !ok || len(fhs) == 0 || fhs[0] == nil {
		return nil
	}
	if len(fhs) > 1 {
		return fmt.Errorf("multiple files uploaded for single file field %s", currentPath)
	}
	if v.CanSet() {
		v.Set(reflect.ValueOf(fhs[0]))
	}
	return nil
}

// fillStruct 遍历结构体字段，递归填充
func fillStruct(v reflect.Value, t reflect.Type, files map[string][]*multipart.FileHeader, currentPath string) error {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		if !field.CanSet() {
			continue
		}

		tagVal := fieldType.Tag.Get("form")
		fieldNameInPath := extractFieldName(tagVal, fieldType.Name)

		// 处理 form:"-" 标签：跳过该字段
		if fieldNameInPath == "-" {
			continue
		}

		newPath := currentPath

		if fieldType.Anonymous {
			// 严格遵循 Go 嵌入规范决定是否提升字段
			isFilePtr := fieldType.Type == reflect.TypeFor[*multipart.FileHeader]()
			isFileSlice := fieldType.Type == reflect.TypeFor[[]*multipart.FileHeader]()

			if isFilePtr || isFileSlice {
				// 文件类型匿名字段：视为普通字段，拼接路径
				if currentPath != "" {
					newPath = currentPath + "." + fieldNameInPath
				} else {
					newPath = fieldNameInPath
				}
			} else {
				// 非文件类型，判断是否应该提升（嵌入结构体 T 或指向结构体的指针 *T）
				typ := fieldType.Type
				shouldPromote := false
				if typ.Kind() == reflect.Struct {
					shouldPromote = true
				} else if typ.Kind() == reflect.Pointer && typ.Elem().Kind() == reflect.Struct {
					shouldPromote = true
				}
				if shouldPromote {
					// 字段提升，路径不变
					newPath = currentPath
				} else {
					// 其他类型（基本类型、切片非结构体等），拼接路径
					if currentPath != "" {
						newPath = currentPath + "." + fieldNameInPath
					} else {
						newPath = fieldNameInPath
					}
				}
			}
		} else {
			// 非匿名字段：正常拼接路径
			if currentPath != "" {
				newPath = currentPath + "." + fieldNameInPath
			} else {
				newPath = fieldNameInPath
			}
		}

		if err := fillFilesRecursive(field, field.Type(), files, newPath); err != nil {
			return err
		}
	}
	return nil
}

// fillSlice 处理元素为结构体或结构体指针的切片
func fillSlice(v reflect.Value, files map[string][]*multipart.FileHeader, currentPath string) error {
	indexFormat := make(map[int]string) // idx -> "bracket" or "dot"

	for key := range files {
		if matches := indexedBracketPattern.FindStringSubmatch(key); matches != nil && matches[1] == currentPath {
			if idx, err := strconv.Atoi(matches[2]); err == nil && idx >= 0 {
				if existing, ok := indexFormat[idx]; ok && existing != "bracket" {
					return fmt.Errorf("index %d in path %s mixes bracket and dot formats (key: %s)", idx, currentPath, key)
				}
				indexFormat[idx] = "bracket"
			}
		}
		if matches := indexedDotPattern.FindStringSubmatch(key); matches != nil && matches[1] == currentPath {
			if idx, err := strconv.Atoi(matches[2]); err == nil && idx >= 0 {
				if existing, ok := indexFormat[idx]; ok && existing != "dot" {
					return fmt.Errorf("index %d in path %s mixes bracket and dot formats (key: %s)", idx, currentPath, key)
				}
				indexFormat[idx] = "dot"
			}
		}
	}

	if len(indexFormat) == 0 {
		return nil
	}

	// 校验整个字段索引格式统一
	var globalFormat string
	for idx, format := range indexFormat {
		if globalFormat == "" {
			globalFormat = format
		} else if globalFormat != format {
			return fmt.Errorf(
				"inconsistent index format for path %s: index %d uses %s, but index 0 uses %s",
				currentPath, idx, format, globalFormat,
			)
		}
	}

	indexes := make([]int, 0, len(indexFormat))
	for idx := range indexFormat {
		indexes = append(indexes, idx)
	}
	sort.Ints(indexes)
	maxIdx := indexes[len(indexes)-1]
	requiredLen := maxIdx + 1

	// 处理 nil 切片：如果切片为 nil，先创建零值切片（后续会扩展）
	if v.IsNil() {
		newSlice := reflect.MakeSlice(v.Type(), 0, 0)
		v.Set(newSlice)
	}

	if v.Len() < requiredLen {
		newSlice := reflect.MakeSlice(v.Type(), requiredLen, requiredLen)
		for i := 0; i < v.Len(); i++ {
			newSlice.Index(i).Set(v.Index(i))
		}
		v.Set(newSlice)
	}

	for _, idx := range indexes {
		if idx >= v.Len() {
			continue
		}
		format := indexFormat[idx]
		var elemPath string
		if format == "bracket" {
			elemPath = fmt.Sprintf("%s[%d]", currentPath, idx)
		} else {
			elemPath = fmt.Sprintf("%s.%d", currentPath, idx)
		}

		elem := v.Index(idx)
		if !elem.CanSet() {
			continue
		}

		switch elem.Kind() {
		case reflect.Struct:
			if err := fillFilesRecursive(elem, elem.Type(), files, elemPath); err != nil {
				return err
			}
		case reflect.Pointer:
			if elem.Type().Elem().Kind() == reflect.Struct {
				if elem.IsNil() {
					// 这里不需要再次检查文件键，因为索引本身已通过 indexFormat 确认有文件存在
					elem.Set(reflect.New(elem.Type().Elem()))
				}
				if err := fillFilesRecursive(elem.Elem(), elem.Elem().Type(), files, elemPath); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
