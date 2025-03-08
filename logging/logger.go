//
// This file is part of the GoBarcodeRelay distribution (https://github.com/SirAfino/go-barcode-relay).
// Copyright (c) 2025 Gabriele Serafino.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.
//

package logging

import (
	"log"
	"os"
)

var logger *log.Logger = log.New(os.Stdout, "", log.LstdFlags)

type Logger struct {
	component string
}

func GetLogger(component string) *Logger {
	return &Logger{
		component: component,
	}
}

func (L *Logger) printf(level string, format string, v ...any) {
	logger.Printf("%s [%s] "+format, append([]any{level, L.component}, v...)...)
}

func (L *Logger) Info(format string, v ...any) {
	L.printf("INFO", format, v...)
}

func (L *Logger) Error(format string, v ...any) {
	L.printf("ERROR", format, v...)
}
