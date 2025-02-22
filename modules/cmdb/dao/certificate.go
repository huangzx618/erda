// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package dao

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateCertificate 创建Certificate
func (client *DBClient) CreateCertificate(certificate *model.Certificate) error {
	return client.Create(certificate).Error
}

// UpdateCertificate 更新Certificate
func (client *DBClient) UpdateCertificate(certificate *model.Certificate) error {
	return client.Save(certificate).Error
}

// DeleteCertificate 删除Certificate
func (client *DBClient) DeleteCertificate(certificateID int64) error {
	return client.Where("id = ?", certificateID).Delete(&model.Certificate{}).Error
}

// GetCertificateByID 根据certificateID获取Certificate信息
func (client *DBClient) GetCertificateByID(certificateID int64) (model.Certificate, error) {
	var certificate model.Certificate
	if err := client.Where("id = ?", certificateID).Find(&certificate).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return certificate, ErrNotFoundCertificate
		}
		return certificate, err
	}
	return certificate, nil
}

// GetCertificatesByOrgIDAndName 根据orgID与名称获取Certificate列表
func (client *DBClient) GetCertificatesByOrgIDAndName(orgID int64, params *apistructs.CertificateListRequest) (
	int, []model.Certificate, error) {
	var (
		certificates []model.Certificate
		total        int
	)
	db := client.Where("org_id = ?", orgID)
	if params.Name != "" {
		db = db.Where("name = ?", params.Name)
	}
	if params.Type != "" {
		db = db.Where("type = ?", params.Type)
	}
	if params.Query != "" {
		db = db.Where("name LIKE ?", strutil.Concat("%", params.Query, "%"))
	}
	db = db.Order("updated_at DESC")
	if err := db.Offset((params.PageNo - 1) * params.PageSize).Limit(params.PageSize).
		Find(&certificates).Error; err != nil {
		return 0, nil, err
	}

	// 获取总量
	db = client.Model(&model.Certificate{}).Where("org_id = ?", orgID)
	if params.Name != "" {
		db = db.Where("name = ?", params.Name)
	}
	if params.Type != "" {
		db = db.Where("type = ?", params.Type)
	}
	if params.Query != "" {
		db = db.Where("name LIKE ?", strutil.Concat("%", params.Query, "%"))
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	return total, certificates, nil
}

// GetCertificatesByIDs 根据certificateIDs获取Certificate列表
func (client *DBClient) GetCertificatesByIDs(certificateIDs []int64, params *apistructs.CertificateListRequest) (
	int, []model.Certificate, error) {
	var (
		total        int
		certificates []model.Certificate
	)
	db := client.Where("id in (?)", certificateIDs)
	if params.Name != "" {
		db = db.Where("name = ?", params.Name)
	}
	if params.Query != "" {
		db = db.Where("name LIKE ?", strutil.Concat("%", params.Query, "%"))
	}
	db = db.Order("updated_at DESC")
	if err := db.Offset((params.PageNo - 1) * params.PageSize).Limit(params.PageSize).
		Find(&certificates).Error; err != nil {
		return 0, nil, err
	}

	// 获取总量
	db = client.Model(&model.Certificate{}).Where("id in (?)", certificateIDs)
	if params.Name != "" {
		db = db.Where("name = ?", params.Name)
	}
	if params.Query != "" {
		db = db.Where("name LIKE ?", strutil.Concat("%", params.Query, "%"))
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	return total, certificates, nil
}

// GetCertificateByOrgAndName 根据orgID & Certificate名称 获取证书信息
func (client *DBClient) GetCertificateByOrgAndName(orgID int64, name string) (*model.Certificate, error) {
	var certificate model.Certificate
	if err := client.Where("org_id = ?", orgID).
		Where("name = ?", name).Find(&certificate).Error; err != nil {
		return nil, err
	}
	return &certificate, nil
}
