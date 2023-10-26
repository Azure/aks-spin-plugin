package spin

import (
	"reflect"
	"testing"
)

func TestNewRuntimeConfig(t *testing.T) {
	type args struct {
		manifest Manifest
	}
	tests := []struct {
		name string
		args args
		want RuntimeConfig
	}{
		{
			name: "no variables",
			args: args{
				manifest: Manifest{
					Variables: map[string]Variable{},
				},
			},
			want: RuntimeConfig{
				Variables: map[string]RuntimeVariable{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRuntimeConfig(tt.args.manifest)
			for k, _ := range got.Variables {
				if !reflect.DeepEqual(got.Variables[k], tt.want.Variables[k]) {
					t.Errorf("NewRuntimeConfig() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
