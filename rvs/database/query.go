package database

import (
	l "altair/rvs/globlog"
	"altair/rvs/utils"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var env = "prod"

func SetupDb(dbUrl string) {

	l.Log().Info("Connecting to the database: ", dbUrl)
	var newLogger logger.Interface

	if env == "dev" {
		newLogger = logger.New(
			l.Log(), // logrus logger
			logger.Config{
				LogLevel:                  logger.Info, // Log level  Info, Error, Warn, Silent
				IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
				Colorful:                  true,        // Disable color
			},
		)
	} else {
		newLogger = logger.New(
			l.Log(), // logrus logger
			logger.Config{
				LogLevel:                  logger.Silent, // Log level  Info, Error, Warn, Silent
				IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
				Colorful:                  true,          // Disable color
			},
		)
	}

	// Connect to the database
	dataBase, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		l.Log().Error("Error while connecting to the database: %v", err)
	} else {
		l.Log().Debug("Successfully connected to the database")
	}

	db = dataBase // Set the database to the global variable

	pool, err := db.DB() // Get the underlying database connection pool
	if err != nil {
		l.Log().Error("Error while connecting to the database: %v", err)
	} else {
		l.Log().Debug("Connection Polling Started")
	}

	pool.SetConnMaxIdleTime(10) // 10 seconds
	pool.SetMaxOpenConns(100)   // 100 connections
	pool.SetMaxIdleConns(100)   // 100 connections
	pool.SetConnMaxLifetime(10) // 10 seconds
}

func SaveTemplate(template Template) error {

	if template.DefaultTemplate {
		UnsetPreviousDefaultTemplate(template.FileExt, template.UserName, template.SeriesFile)
	}

	err := db.Table("template_data_store_table").Create(&template).Error

	if err != nil {
		return err
	} else {
		l.Log().Info("template inserted in the database: ", template.TemplateId)
	}

	return nil

}

func GetTemplateDetails(templateId string) (Template, error) {

	var templateData Template
	err := db.Table("template_data_store_table").Where("template_id_col = ?", templateId).Find(&templateData).Error
	if err != nil {
		return Template{}, err
	} else {
		l.Log().Info("Template details fetched successfully: ", templateData.TemplateId)
	}
	return templateData, nil
}

func SetTemplateAsDefaultValue(sTemplateId string, fileextension string, isSeriesFile bool, username string) (bool, error) {

	UnsetPreviousDefaultTemplate(fileextension, username, isSeriesFile)

	if utils.IsValidString(sTemplateId) {
		err := db.Exec("UPDATE template_data_store_table SET default_template_col = ? WHERE template_id_col = ?",
			true, sTemplateId).Error

		if err != nil {
			return false, err
		} else {
			l.Log().Info("Template set to default successfully: ", sTemplateId)
		}
	}
	return true, nil

}

func DeleteTemplateData(sTemplateId string) (bool, error) {

	fmt.Println("Delete sTemplateId ", sTemplateId)
	err := db.Exec("DELETE FROM template_data_store_table WHERE template_id_col = ?", sTemplateId).Error
	if err != nil {
		return false, err
	} else {
		l.Log().Debug("Selected Template deleted from the database: ", sTemplateId)
	}
	return true, nil
}

func UpdateTemplateData(template Template) (bool, error) {

	// if Template is already in the db, update it else insert it in the db
	var templateData Template
	if template.DefaultTemplate {
		UnsetPreviousDefaultTemplate(template.FileExt, template.UserName, template.SeriesFile)
	}

	err := db.Table("template_data_store_table").Where("template_id_col = ?", template.TemplateId).First(&templateData).Error
	if err != nil {
		err := db.Table("template_data_store_table").Create(&template).Error
		if err != nil {
			return false, err
		} else {
			l.Log().Info("Template inserted in the database: ", template.TemplateId)
		}
	} else {
		err := db.Table("template_data_store_table").Where("template_id_col = ?", template.TemplateId).Updates(template).Error
		if err != nil {
			return false, err
		} else {
			l.Log().Info("Template updated in the database: ", template.TemplateId)
		}
	}

	return true, nil
}

/**
 * This method unset previous default template.
 *
 * @param tocRequestForResult - plot toc request
 * @param plotTOCOutputFile - original plot toc file
 * @param filteredTOCFile - partial plot toc file
 * @param userCredentials - user credentials
 */
func UnsetPreviousDefaultTemplate(fileextension string, username string, isSeriesFile bool) error {
	l.Log().Debug("Entering method unsetPreviousDefaultTemplate")
	var listOfTemplates []Template

	l.Log().Info("getting all templates")

	err := db.Table("template_data_store_table").Where("file_extension_col = ? AND user_name_col = ? AND default_template_col = ? AND series_file_col = ?",
		fileextension, username, true, isSeriesFile).Find(&listOfTemplates).Error
	if err != nil {
		l.Log().Error("Error while fetching templates from the database: ", err)
	}

	if len(listOfTemplates) != 0 {
		err := db.Exec("UPDATE template_data_store_table SET default_template_col = ? WHERE template_id_col = ?",
			false, listOfTemplates[0].TemplateId).Error
		if err != nil {
			return err
		} else {
			l.Log().Debug("template_data_store_table updated in the database: ")
		}
		return nil
	}

	return nil
}

func GetAllTemplateData(fileextension string, username string, isSeriesFile bool) ([]Template, error) {

	var listOfTemplates []Template
	err := db.Table("template_data_store_table").Where("file_extension_col = ? AND user_name_col = ? AND series_file_col = ?",
		fileextension, username, isSeriesFile).Find(&listOfTemplates).Error
	if err != nil {
		return nil, err
	} else {
		l.Log().Info("All Template details fetched successfully: ")
	}
	return listOfTemplates, nil
}
