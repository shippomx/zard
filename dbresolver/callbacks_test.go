package dbresolver

import "testing"

func Test_isWITHSelect(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want bool
	}{
		{
			name: "simple WITH SELECT",
			sql:  "WITH cte AS (SELECT * FROM users) SELECT * FROM cte",
			want: true,
		},
		{
			name: "complex WITH SELECT",
			sql:  "WITH user_stats AS (SELECT department, COUNT(*) as count FROM users GROUP BY department) SELECT * FROM user_stats",
			want: true,
		},
		{
			name: "lowercase with select",
			sql:  "with my_cte as (select 1) select * from my_cte",
			want: true,
		},
		{
			name: "mixed case WITH SELECT",
			sql:  "WiTh test_cte AS (SELECT id FROM table) SELECT * FROM test_cte",
			want: true,
		},
		{
			name: "WITH UPDATE - should not match",
			sql:  "WITH cte AS (SELECT * FROM users) UPDATE table SET col = 1",
			want: false,
		},
		{
			name: "WITH DELETE - should not match",
			sql:  "WITH old_users AS (SELECT id FROM users) DELETE FROM users",
			want: false,
		},
		{
			name: "regular SELECT statement",
			sql:  "SELECT * FROM users",
			want: false,
		},
		{
			name: "WITH in table name",
			sql:  "SELECT * FROM table_with_stats",
			want: false,
		},
		{
			name: "empty string",
			sql:  "",
			want: false,
		},
		{
			name: "invalid WITH clause",
			sql:  "WITH AS (SELECT * FROM users) SELECT *",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isWITHSelect(tt.sql); got != tt.want {
				t.Errorf("isWITHSelect() = %v, want %v", got, tt.want)
			}
		})
	}
}
