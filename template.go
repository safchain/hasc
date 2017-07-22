/*
 * Copyright (C) 2017 Sylvain Afchain
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

package hasc

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/safchain/hasc/statics"
)

func itemTemplate(tmplPath string, data interface{}) template.HTML {
	asset := statics.MustAsset(tmplPath)

	tmpl := template.Must(template.New("tmpl").Parse(string(asset)))

	var result bytes.Buffer
	tmpl.Execute(&result, data)

	return template.HTML(result.String())
}

func valueTemplate(i Item, label string, unit string, img string) template.HTML {
	data := struct {
		ID        string
		Label     string
		ItemLabel string
		Unit      string
		Img       string
	}{
		ID:        i.Object().ID() + "_" + i.ID(),
		Label:     i.Object().Label(),
		ItemLabel: label,
		Unit:      unit,
		Img:       fmt.Sprintf("statics/img/%s.png", img),
	}

	return itemTemplate("statics/items/value.html", data)
}
