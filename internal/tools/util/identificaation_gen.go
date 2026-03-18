package util

import (
	"fmt"
	"strings"
)

const (
	ID_PREFIX_QUOTATION = "QUO"
	ID_PREFIX_INVOICE   = "INV"
	ID_PREFIX_RECEIPT   = "REC"
)

func getLastId() string {

	return ""
}

func IdentificationGenerator(uniqueId string, documentType string) string {
	lastId := strings.Split(getLastId(), "/")

	generated_id := fmt.Sprintf("%s-%s-%s", documentType, uniqueId, lastId)

	return generated_id
}
