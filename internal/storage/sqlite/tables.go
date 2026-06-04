package sqlite

import (
	"fmt"

	"github.com/fastygo/app-gocms/pkg/module/records"
	"github.com/fastygo/platform/pkg/toolset"
)

const (
	RecordSetting toolset.RecordTypeID = "setting"
	RecordMenu    toolset.RecordTypeID = "menu"
)

type recordTable struct {
	RecordType string
	Name       string
}

func cmsTables() []recordTable {
	return []recordTable{
		{RecordType: string(records.RecordPost), Name: "posts"},
		{RecordType: string(records.RecordPage), Name: "pages"},
		{RecordType: string(records.RecordContentType), Name: "content_types"},
		{RecordType: string(records.RecordContentMeta), Name: "content_meta_definitions"},
		{RecordType: string(records.RecordTaxonomy), Name: "taxonomies"},
		{RecordType: string(records.RecordTerm), Name: "terms"},
		{RecordType: string(records.RecordMediaAsset), Name: "media_assets"},
		{RecordType: string(records.RecordAuthor), Name: "authors"},
		{RecordType: string(RecordSetting), Name: "settings"},
		{RecordType: string(RecordMenu), Name: "menus"},
	}
}

func tableForRecordType(recordType string) (recordTable, error) {
	for _, table := range cmsTables() {
		if table.RecordType == recordType {
			return table, nil
		}
	}
	return recordTable{}, fmt.Errorf("unsupported AppCMS record type %q", recordType)
}
