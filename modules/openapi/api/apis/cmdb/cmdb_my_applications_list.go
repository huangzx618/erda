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

package cmdb

import (
	"github.com/erda-project/erda/apistructs"

	"github.com/erda-project/erda/modules/openapi/api/apis"
)

/**
add for standard
*/
var CMDB_MY_APPLICATIONS_LIST = apis.ApiSpec{
	Path:         "/api/applications/actions/list-my-applications",
	BackendPath:  "/api/applications/actions/list-my-applications",
	Host:         "cmdb.marathon.l4lb.thisdcos.directory:9093",
	Scheme:       "http",
	Method:       "GET",
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.ApplicationListRequest{},
	ResponseType: apistructs.ApplicationListResponse{},
	IsOpenAPI:    true,
	Doc:          "summary: 查询我的应用列表",
}
