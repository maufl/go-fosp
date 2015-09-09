// Copyright (C) 2015 Felix Maurer
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package main

import (
	"io/ioutil"
	"log"
	"time"
)

func indexOf(array []string, element string) int {
	for i, v := range array {
		if v == element {
			return i
		}
	}
	return -1
}

func contains(array []string, element string) bool {
	return indexOf(array, element) != -1
}

var performanceLogger *log.Logger

func init() {
	if performanceLogger != nil {
		return
	}
	println("Creating new log file")
	timeString := time.Now().Format("2006-02-01_15:04")
	if file, err := ioutil.TempFile("", "fosp-perf-log-"+timeString+"-"); err == nil {
		performanceLogger = log.New(file, "", 0)
	} else {
		println("Could not open performance log")
	}
}

func timeTrack(start time.Time, name string) {
	if performanceLogger == nil {
		return
	}
	elapsed := time.Since(start)
	performanceLogger.Printf("%s took %s", name, elapsed)
}
