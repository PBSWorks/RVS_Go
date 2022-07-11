package database

type Template struct {
	TemplateId          int64  `gorm:"primary_key;column:template_id_col"`
	TemplateName        string `gorm:"column:template_name_col;not null"`
	FileName            string `gorm:"column:file_name_col;not null"`
	FileExt             string `gorm:"column:file_extension_col;not null"`
	SeriesFile          bool   `gorm:"column:series_file_col"`
	ApplicationName     string `gorm:"column:application_name_col"`
	UserName            string `gorm:"column:user_name_col;not null"`
	DefaultTemplate     bool   `gorm:"column:default_template_col"`
	TemplateData        []byte `gorm:"column:template_data_col;not null"`
	FilteredReqTemplate bool   `gorm:"column:filter_req_template_col"`
}
