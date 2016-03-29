package docker

import (
	"strconv"
	"strings"
)

func ImageEquals(a, b interface{}) bool {
	imageA, ok := a.(string)
	if !ok {
		return false
	}
	imageB, ok := b.(string)
	if !ok {
		return false
	}
	return imageA == imageB
}

func ImageLessBySemanticVersion(a, b interface{}) bool {
	imageA, ok := a.(string)
	if !ok {
		return false
	}

	imageB, ok := b.(string)
	if !ok {
		return false
	}

	repoA, tagA, err := ParseDockerImage(imageA)
	if err != nil {
		return false
	}
	repoB, tagB, err := ParseDockerImage(imageB)
	if err != nil {
		return false
	}

	if repoA != repoB {
		return false
	}

	sep := func(c rune) bool { return c == '.' || c == ',' || c == '-' }
	min := func(a, b int) int {
		if a < b {
			return a
		} else {
			return b
		}
	}

	// compare the tags... we tokenize the tags by delimiters such as . and -
	fieldsA := strings.FieldsFunc(tagA, sep)
	fieldsB := strings.FieldsFunc(tagB, sep)
	for i := 0; i < min(len(fieldsA), len(fieldsB)); i++ {
		a, erra := strconv.Atoi(fieldsA[i])
		b, errb := strconv.Atoi(fieldsB[i])
		switch {
		case erra != nil && errb != nil:
			if fieldsA[i] == fieldsB[i] {
				continue
			} else {
				return fieldsA[i] < fieldsB[i]
			}
		case erra == nil && errb == nil:
			if a == b {
				continue
			} else {
				return a < b
			}
		case erra != nil || errb != nil:
			return false
		}
	}
	return false
}
