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

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTruncateRunesWithSuffix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		maxRunes int
		suffix   string
		want     string
	}{
		{
			name:     "no truncation",
			input:    "active",
			maxRunes: 50,
			suffix:   "...",
			want:     "active",
		},
		{
			name:     "truncate ascii",
			input:    "teleport is starting and has not joined",
			maxRunes: 20,
			suffix:   "...",
			want:     "teleport is start...",
		},
		{
			name:     "truncate unicode runes",
			input:    "áéíóú",
			maxRunes: 4,
			suffix:   "...",
			want:     "á...",
		},
		{
			name:     "suffix rune aware",
			input:    "abcdef",
			maxRunes: 5,
			suffix:   "⚠️",
			want:     "abc⚠️",
		},
		{
			name:     "max smaller than suffix",
			input:    "teleport",
			maxRunes: 1,
			suffix:   "...",
			want:     "...",
		},
		{
			name:     "empty input",
			input:    "",
			maxRunes: 1,
			suffix:   "...",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, TruncateRunesWithSuffix(tt.input, tt.maxRunes, tt.suffix))
		})
	}
}
