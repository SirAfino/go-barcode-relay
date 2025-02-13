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

package reader

var CharMap = map[uint16]string{
	2:  "1",
	3:  "2",
	4:  "3",
	5:  "4",
	6:  "5",
	7:  "6",
	8:  "7",
	9:  "8",
	10: "9",
	11: "0",
	16: "Q",
	17: "W",
	18: "E",
	19: "R",
	20: "T",
	21: "Y",
	22: "U",
	23: "I",
	24: "O",
	25: "P",
	28: "\n", // Enter
	30: "A",
	31: "S",
	32: "D",
	33: "F",
	34: "G",
	35: "H",
	36: "J",
	37: "K",
	38: "L",
	42: "", // Shift
	44: "Z",
	45: "X",
	46: "C",
	47: "V",
	48: "B",
	49: "N",
	50: "M",
}
