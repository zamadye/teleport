/*
 * Teleport
 * Copyright (C) 2026  Gravitational, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package utils

import "unicode/utf8"

// TruncateRunesWithSuffix truncates s to at most maxRunes runes, appending
// suffix if truncation occurred.
func TruncateRunesWithSuffix(s string, maxRunes int, suffix string) string {
	limit := max(maxRunes-utf8.RuneCountInString(suffix), 0)

	i := 0
	for pos := range s {
		if i >= limit {
			return s[:pos] + suffix
		}
		i++
	}

	return s
}
