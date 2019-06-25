package helpers

import "bytes"

/*
  StringBuilder struct.
	  Usage:
		builder := NewStringBuilder()
		builder.Append("a").Append("b").Append("c")
		s := builder.String()
		println(s)
	  print:
		abc
*/

// StringBuilder ...
type StringBuilder struct {
	buffer bytes.Buffer
}

// NewStringBuilder ...
func NewStringBuilder() *StringBuilder {
	var builder StringBuilder
	return &builder
}

// Append ...
func (builder *StringBuilder) Append(s string) *StringBuilder {
	builder.buffer.WriteString(s)
	return builder
}

// AppendStrings ...
func (builder *StringBuilder) AppendStrings(ss ...string) *StringBuilder {
	for i := range ss {
		builder.buffer.WriteString(ss[i])
	}
	return builder
}

// Clear ...
func (builder *StringBuilder) Clear() *StringBuilder {
	var buffer bytes.Buffer
	builder.buffer = buffer
	return builder
}

// ToString ...
func (builder *StringBuilder) ToString() string {
	return builder.buffer.String()
}

//ComparisonSlieString 去重,比较切片
func ComparisonSlieString(news []string, olds []string) []string {
	var oslie []string
	temp := map[string]struct{}{}
	for _, v := range news {
		if _, ok := temp[v]; !ok {
			temp[v] = struct{}{}
			oslie = append(oslie, v)
		}
	}
	news = oslie
	oslie = oslie[:0]
	for _, vo := range news {
		var is bool
		for _, vn := range olds {
			if vo == vn {
				is = true
			}
		}
		if !is {
			oslie = append(oslie, vo)
		}
	}
	return oslie
}
