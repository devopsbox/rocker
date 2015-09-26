/*-
 * Copyright 2015 Grammarly, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package build2

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
)

type Cache interface {
	Get(s State) (s2 *State, err error)
	Put(s State) error
}

type CacheFS struct {
	root string
}

func NewCacheFS(root string) *CacheFS {
	return &CacheFS{
		root: root,
	}
}

func (c *CacheFS) Get(s State) (res *State, err error) {
	match := filepath.Join(c.root, s.ImageID)

	err = filepath.Walk(match, func(path string, info os.FileInfo, err error) error {
		if err != nil && os.IsNotExist(err) {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		s2 := State{}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &s2); err != nil {
			return err
		}

		log.Debugf("CACHE COMPARE %s %s %q %q", s.ImageID, s2.ImageID, s.Commits, s2.Commits)

		if s.Equals(s2) {
			res = &s2
			return filepath.SkipDir
		}
		return nil
	})

	if err == filepath.SkipDir {
		return res, nil
	}

	return
}

func (c *CacheFS) Put(s State) error {
	log.Debugf("CACHE PUT %s %s %q", s.ParentID, s.ImageID, s.Commits)

	fileName := filepath.Join(c.root, s.ParentID, s.ImageID) + ".json"
	if err := os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fileName, data, 0644)
}
