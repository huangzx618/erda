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

// +build !codeanalysis

package gitmodule

import (
	"errors"

	git "github.com/libgit2/git2go/v30"
)

type EditAction string
type EditPathType string

const (
	EDIT_ACTION_ADD    EditAction = "add"
	EDIT_ACTION_DELETE EditAction = "delete"
	EDIT_ACTION_MOVE   EditAction = "move"
)

const (
	EDIT_PATH_TYPE_TREE EditPathType = "tree"
	EDIT_PATH_TYPE_BLOB EditPathType = "blob"
)

type EditActionItem struct {
	Action   EditAction   `json:"action"`
	Content  string       `json:"content"`
	Path     string       `json:"path"`
	PathType EditPathType `json:"pathType"`
}

type CreateCommit struct {
	Signature *Signature        `json:"-"`
	Message   string            `json:"message"`
	Actions   []*EditActionItem `json:"actions"`
	Branch    string            `json:"branch"`
}

func (repo *Repository) CreateCommit(request *CreateCommit) (*Commit, error) {
	branch := request.Branch
	message := request.Message
	isInitCommit := false
	rawRepo, err := repo.GetRawRepo()
	if err != nil {
		return nil, err
	}
	isEmpty, err := rawRepo.IsEmpty()
	if err != nil {
		return nil, err
	}
	if isEmpty {
		isInitCommit = true
	}

	index, err := rawRepo.Index()
	defer index.Free()
	if err != nil {
		return nil, err
	}

	parentCommits := []*git.Commit{}
	if !isInitCommit {
		// 不是init commit,读取对应分支内容到index
		branchCommit, err := repo.GetBranchCommit(branch)
		if err != nil {
			return nil, err
		}
		oldTreeOid, _ := git.NewOid(branchCommit.TreeSha)
		oldTree, err := rawRepo.LookupTree(oldTreeOid)
		if err != nil {
			return nil, err
		}
		parentOid, _ := git.NewOid(branchCommit.ID)
		parentCommit, err := rawRepo.LookupCommit(parentOid)
		if err != nil {
			return nil, err
		}
		parentCommits = append(parentCommits, parentCommit)
		index.ReadTree(oldTree)
	}

	for _, action := range request.Actions {
		if action.PathType == "" {
			action.PathType = EDIT_PATH_TYPE_BLOB
		}
	}

	for _, action := range request.Actions {
		if action.Action == EDIT_ACTION_ADD {
			if action.PathType == EDIT_PATH_TYPE_TREE {
				oid, err := rawRepo.CreateBlobFromBuffer([]byte(""))
				if err != nil {
					return nil, err
				}

				index.Add(&git.IndexEntry{
					Mode: git.FilemodeBlob,
					Id:   oid,
					Path: action.Path + "/.gitkeep",
				})

			} else if action.PathType == EDIT_PATH_TYPE_BLOB {
				content := action.Content
				oid, err := rawRepo.CreateBlobFromBuffer([]byte(content))
				if err != nil {
					return nil, err
				}

				index.Add(&git.IndexEntry{
					Mode: git.FilemodeBlob,
					Id:   oid,
					Path: action.Path,
				})

			} else {
				return nil, errors.New("not support path type: " + string(action.PathType))
			}
		} else if action.Action == EDIT_ACTION_DELETE {
			if action.PathType == EDIT_PATH_TYPE_TREE {
				index.RemoveDirectory(action.Path, 0)
			} else if action.PathType == EDIT_PATH_TYPE_BLOB {
				index.RemoveByPath(action.Path)
			} else {
				return nil, errors.New("not support pathType: " + string(action.PathType))
			}
		} else {
			return nil, errors.New("not support action: " + string(action.Action))
		}
	}

	err = index.Write()
	if err != nil {
		return nil, err
	}

	sig := &git.Signature{
		Name:  request.Signature.Name,
		Email: request.Signature.Email,
		When:  request.Signature.When,
	}

	newTreeOid, err := index.WriteTree()
	if err != nil {
		return nil, err
	}

	newTree, err := rawRepo.LookupTree(newTreeOid)
	if err != nil {
		return nil, err
	}

	newOid, err := rawRepo.CreateCommit(BRANCH_PREFIX+branch, sig, sig, message, newTree, parentCommits...)
	if err != nil {
		return nil, err
	}
	if isInitCommit {
		//把第一次提交的分支设为默认分支
		err := rawRepo.SetHead(BRANCH_PREFIX + branch)
		if err != nil {
			return nil, err
		}
	}

	return repo.GetCommit(newOid.String())
}
