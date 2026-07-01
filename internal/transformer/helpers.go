package transformer

import (
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/kmmz532/php2js/internal/jsast"
)

// extractVarName gets the variable name from a variable node.
func (t *Transformer) extractVarName(node ast.Vertex) string {
	switch n := node.(type) {
	case *ast.ExprVariable:
		if id, ok := n.Name.(*ast.Identifier); ok {
			return sanitizeVarName(string(id.Value))
		}
		if name := t.extractName(n.Name); name != "unknown" {
			return sanitizeVarName(name)
		}
	case *ast.Identifier:
		return sanitizeVarName(string(n.Value))
	}
	return "unknown"
}

// extractName gets a name from an identifier node.
func (t *Transformer) extractName(node ast.Vertex) string {
	switch n := node.(type) {
	case *ast.Identifier:
		return string(n.Value)
	case *ast.Name:
		return t.nameToString(n)
	case *ast.NameFullyQualified:
		return t.namePartsToString(n.Parts)
	}
	return "unknown"
}

// extractCallName extracts the function/class name from a call target.
func (t *Transformer) extractCallName(node ast.Vertex) string {
	switch n := node.(type) {
	case *ast.Name:
		return t.nameToString(n)
	case *ast.NameFullyQualified:
		return t.namePartsToString(n.Parts)
	case *ast.Identifier:
		return string(n.Value)
	case *ast.ExprVariable:
		return t.extractVarName(n)
	}
	return "unknown"
}

// nameToString converts a Name node to a string.
func (t *Transformer) nameToString(n *ast.Name) string {
	return t.namePartsToString(n.Parts)
}

// namePartsToString joins name parts into a string.
func (t *Transformer) namePartsToString(parts []ast.Vertex) string {
	var names []string
	for _, p := range parts {
		if np, ok := p.(*ast.NamePart); ok {
			names = append(names, string(np.Value))
		}
	}
	return strings.Join(names, "_")
}

// getModifiers extracts access modifier string.
func (t *Transformer) getModifiers(mods []ast.Vertex) string {
	for _, m := range mods {
		if id, ok := m.(*ast.Identifier); ok {
			v := string(id.Value)
			if v == "private" || v == "protected" || v == "public" || v == "static" {
				return v
			}
		}
	}
	return "public"
}

// sanitizeVarName makes a PHP variable name safe for JS.
func sanitizeVarName(name string) string {
	// Remove $ prefix if present
	name = strings.TrimPrefix(name, "$")
	if jsReservedWords[name] {
		return name + "_"
	}
	// Replace invalid chars
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

// escapeJSString escapes a string for JS.
func escapeJSString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "$", "\\$")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// mapFunctionName maps PHP function names to JS runtime equivalents.
func (t *Transformer) mapFunctionName(name string) jsast.Expression {
	// PHP built-in functions -> runtime
	builtins := map[string]string{
		// String functions
		"strlen": "strlen", "strpos": "strpos", "strrpos": "strrpos",
		"substr": "substr", "str_replace": "str_replace",
		"explode": "explode", "implode": "implode",
		"trim": "trim", "ltrim": "ltrim", "rtrim": "rtrim",
		"strtolower": "strtolower", "strtoupper": "strtoupper",
		"ucfirst": "ucfirst", "lcfirst": "lcfirst",
		"sprintf": "sprintf", "printf": "printf",
		"htmlspecialchars": "htmlspecialchars", "htmlspecialchars_decode": "htmlspecialchars_decode",
		"htmlentities": "htmlentities", "strip_tags": "strip_tags",
		"nl2br": "nl2br", "wordwrap": "wordwrap",
		"str_pad": "str_pad", "str_repeat": "str_repeat",
		"str_split": "str_split", "chunk_split": "chunk_split",
		"ord": "ord", "chr": "chr",
		"mb_strlen": "mb_strlen", "mb_strpos": "mb_strpos",
		"mb_substr": "mb_substr", "mb_strtolower": "mb_strtolower",
		"mb_strtoupper": "mb_strtoupper", "mb_convert_encoding": "mb_convert_encoding",
		"mb_detect_encoding": "mb_detect_encoding", "mb_internal_encoding": "mb_internal_encoding",
		"mb_language": "mb_language", "mb_regex_encoding": "mb_regex_encoding",
		"mb_convert_kana": "mb_convert_kana", "mb_ereg": "mb_ereg",
		"mb_ereg_replace": "mb_ereg_replace", "mb_http_output": "mb_http_output",
		"mb_detect_order": "mb_detect_order",
		// Array functions
		"count": "count", "sizeof": "count",
		"array_push": "array_push", "array_pop": "array_pop",
		"array_shift": "array_shift", "array_unshift": "array_unshift",
		"array_merge": "array_merge", "array_slice": "array_slice",
		"array_splice": "array_splice",
		"array_keys": "array_keys", "array_values": "array_values",
		"array_flip": "array_flip", "array_reverse": "array_reverse",
		"array_map": "array_map", "array_filter": "array_filter",
		"array_reduce": "array_reduce", "array_walk": "array_walk",
		"array_search": "array_search", "in_array": "in_array",
		"array_key_exists": "array_key_exists", "array_unique": "array_unique",
		"sort": "sort", "rsort": "rsort", "asort": "asort",
		"arsort": "arsort", "ksort": "ksort", "krsort": "krsort",
		"usort": "usort", "uksort": "uksort", "uasort": "uasort",
		"array_diff": "array_diff", "array_intersect": "array_intersect",
		"range": "range", "compact": "compact", "extract": "extract",
		"list": "list",
		// Type functions
		"is_array": "is_array", "is_string": "is_string",
		"is_int": "is_int", "is_integer": "is_int",
		"is_float": "is_float", "is_double": "is_float",
		"is_bool": "is_bool", "is_null": "is_null",
		"is_numeric": "is_numeric", "is_object": "is_object",
		"gettype": "gettype", "settype": "settype",
		"intval": "intval", "floatval": "floatval", "strval": "strval",
		// Math
		"abs": "abs", "ceil": "ceil", "floor": "floor", "round": "round",
		"max": "max", "min": "min", "pow": "pow", "sqrt": "sqrt",
		"rand": "rand", "mt_rand": "mt_rand",
		// Hash/Crypto
		"md5": "md5", "sha1": "sha1", "hash": "hash",
		"crc32": "crc32", "base64_encode": "base64_encode",
		"base64_decode": "base64_decode",
		// JSON
		"json_encode": "json_encode", "json_decode": "json_decode",
		// URL
		"urlencode": "urlencode", "urldecode": "urldecode",
		"rawurlencode": "rawurlencode", "rawurldecode": "rawurldecode",
		"http_build_query": "http_build_query", "parse_url": "parse_url",
		"parse_str": "parse_str",
		// Date/Time
		"date": "date", "time": "time", "mktime": "mktime",
		"strtotime": "strtotime", "microtime": "microtime",
		"gmdate": "gmdate",
		// File I/O -> R2
		"file_get_contents": "file_get_contents", "file_put_contents": "file_put_contents",
		"file_exists": "file_exists", "unlink": "unlink", "rename": "rename",
		"fopen": "fopen", "fclose": "fclose", "fread": "fread",
		"fwrite": "fwrite", "fgets": "fgets", "feof": "feof",
		"scandir": "scandir", "glob": "glob",
		"mkdir": "mkdir", "rmdir": "rmdir", "is_dir": "is_dir", "is_file": "is_file",
		"dirname": "dirname", "basename": "basename", "pathinfo": "pathinfo",
		"realpath": "realpath", "file": "file",
		// Output
		"echo": "echo", "print": "echo", "var_dump": "var_dump",
		"print_r": "print_r", "var_export": "var_export",
		// Regex
		"preg_match": "preg_match", "preg_match_all": "preg_match_all",
		"preg_replace": "preg_replace", "preg_split": "preg_split",
		"preg_replace_callback": "preg_replace_callback",
		// HTTP
		"header": "header", "setcookie": "setcookie",
		"session_start": "session_start", "session_destroy": "session_destroy",
		"session_id": "session_id",
		// Misc
		"isset": "isset", "unset": "unset", "empty": "empty",
		"die": "die", "exit": "exit",
		"defined": "defined", "define": "define",
		"class_exists": "class_exists", "function_exists": "function_exists",
		"call_user_func": "call_user_func", "call_user_func_array": "call_user_func_array",
		"sleep": "sleep", "usleep": "usleep",
		// Environment / Error
		"error_reporting": "error_reporting", "ini_set": "ini_set", "ini_get": "ini_get",
		"set_time_limit": "set_time_limit", "memory_get_usage": "memory_get_usage",
		"extension_loaded": "extension_loaded", "get_magic_quotes_gpc": "get_magic_quotes_gpc",
		"constant": "constant", "is_readable": "is_readable", "is_writable": "is_writable",
		"is_callable": "is_callable",
	}

	if jsName, ok := builtins[name]; ok {
		return &jsast.MemberExpr{
			Object:   &jsast.Identifier{Name: "__runtime"},
			Property: &jsast.Identifier{Name: jsName},
		}
	}

	// User-defined function - use as-is
	return &jsast.Identifier{Name: name}
}
