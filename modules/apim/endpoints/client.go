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

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/apim/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
)

// 创建一个客户端
func (e *Endpoints) CreateClient(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.CreateClient.NotLogin().ToResp(), nil
	}

	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.CreateClient.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var body apistructs.CreateClientBody
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		return apierrors.CreateClient.InvalidParameter(err).ToResp(), nil
	}

	var req = apistructs.CreateClientReq{
		OrgID:    orgID,
		Identity: &identity,
		Body:     &body,
	}

	model, apiError := e.assetSvc.CreateClient(&req)
	if apiError != nil {
		return apiError.ToResp(), nil
	}
	return httpserver.OkResp(map[string]interface{}{"client": model})
}

// 获取本人创建的客户端列表
func (e *Endpoints) ListMyClients(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ListClients.NotLogin().ToResp(), nil
	}

	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.ListClients.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var queryParams apistructs.ListMyClientsQueryParams
	if err = e.queryStringDecoder.Decode(&queryParams, r.URL.Query()); err != nil {
		return apierrors.ListClients.InvalidParameter(err).ToResp(), nil
	}

	var req = apistructs.ListMyClientsReq{
		OrgID:       orgID,
		Identity:    &identity,
		QueryParams: &queryParams,
	}
	data, apiError := e.assetSvc.ListMyClients(&req)
	if apiError != nil {
		return apierrors.ListClients.InternalError(apiError).ToResp(), nil
	}

	return httpserver.OkResp(data)
}

// 查询一个客户端的详情
func (e *Endpoints) GetClient(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.CreateClient.NotLogin().ToResp(), nil
	}

	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.CreateClient.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var req = apistructs.GetClientReq{
		OrgID:     orgID,
		Identity:  &identity,
		URIParams: &apistructs.GetClientURIParams{ClientID: vars[urlPathClientID]},
	}

	data, apiError := e.assetSvc.GetMyClient(&req)
	if apiError != nil {
		return apiError.ToResp(), nil
	}

	return httpserver.OkResp(data)
}

func (e *Endpoints) ListSwaggerClient(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ListSwaggerClients.NotLogin().ToResp(), nil
	}

	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.ListSwaggerClients.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var queryParams apistructs.ListSwaggerVersionClientQueryParams
	if err = e.queryStringDecoder.Decode(&queryParams, r.URL.Query()); err != nil {
		return apierrors.ListSwaggerClients.InvalidParameter("invalid query parameters").ToResp(), nil
	}

	var req = apistructs.ListSwaggerVersionClientsReq{
		OrgID:    orgID,
		Identity: &identity,
		URIParams: &apistructs.ListSwaggerVersionClientURIParams{
			AssetID:        vars[urlPathAssetID],
			SwaggerVersion: vars[urlPathSwaggerVersion],
		},
		QueryParams: &queryParams,
	}

	data, apiError := e.assetSvc.ListSwaggerVersionClients(&req)
	if apiError != nil {
		return apiError.ToResp(), nil
	}

	return httpserver.OkResp(data)
}

func (e *Endpoints) UpdateClient(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.UpdateClient.NotLogin().ToResp(), nil
	}

	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.UpdateClient.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	clientModelID, err := strconv.ParseUint(vars[urlPathClientID], 10, 64)
	if err != nil {
		return apierrors.DeleteClient.InvalidParameter("invalid client primary id").ToResp(), nil
	}
	var req = apistructs.UpdateClientReq{
		OrgID:       orgID,
		Identity:    &identity,
		URIParams:   &apistructs.UpdateClientURIParams{ClientID: clientModelID},
		QueryParams: new(apistructs.UpdateClientQueryParams),
		Body:        new(apistructs.UpdateClientBody),
	}

	if err = e.queryStringDecoder.Decode(req.QueryParams, r.URL.Query()); err != nil {
		return apierrors.UpdateClient.InvalidParameter("invalid query parameters").ToResp(), nil
	}
	if err = json.NewDecoder(r.Body).Decode(req.Body); err != nil {
		return apierrors.UpdateClient.InvalidParameter("invalid body").ToResp(), nil
	}

	client, sk, apiError := e.assetSvc.UpdateClient(&req)
	if apiError != nil {
		return apiError.ToResp(), nil
	}

	return httpserver.OkResp(map[string]interface{}{"client": client, "sk": sk})
}

func (e *Endpoints) DeleteClient(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.DeleteClient.NotLogin().ToResp(), nil
	}

	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.DeleteClient.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	clientModelID, err := strconv.ParseUint(vars[urlPathClientID], 10, 64)
	if err != nil {
		return apierrors.DeleteClient.InvalidParameter("invalid client primary id").ToResp(), nil
	}
	var req = apistructs.DeleteClientReq{
		OrgID:    orgID,
		Identity: &identity,
		URIParams: &apistructs.DeleteClientURIParams{
			ClientID: clientModelID,
		},
	}

	if apiError := e.assetSvc.DeleteClient(&req); apiError != nil {
		return apiError.ToResp(), nil
	}

	return httpserver.OkResp(nil)
}
