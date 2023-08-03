/* MIT License

Copyright (c) 2023 Asaduzzaman Pavel (contact@iampavel.dev)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package utils

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func Mkdir(dir string) (string, func()) {
	var err error
	if dir == "" {
		dir, err = os.MkdirTemp("", "jb")
		if err != nil {
			log.Error().Err(err).Msg("Failed to create temp dir")
			os.Exit(1)
		}
		return dir, func() {
			os.RemoveAll(dir)
		}
	} else {
		dir, err = filepath.Abs(dir)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get absolute path")
			os.Exit(1)
		}

		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create dir")
			os.Exit(1)
		}

		// Do not remove the directory if it was provided by the user
		return dir, func() {}
	}
}
