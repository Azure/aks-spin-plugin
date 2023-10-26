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

func Test_loadRuntimeConfig(t *testing.T) {
	type args struct {
		runtimeConfigContents []byte
	}
	tests := []struct {
		name    string
		args    args
		want    RuntimeConfig
		wantErr bool
	}{
		{
			name: "basic runtime config",
			args: args{
				runtimeConfigContents: []byte(`
				 [variables]
					  [variables.foo]
					  default = "bar"
					  required = true	
				 [key_value_store]
					  [key_value_store.foo]
					  type = "azure_cosmos"
					  key = "test-key"
					  account = "test-account"
					  database = "test-db"
					  container = "test-container"
				 `),
			},
			want: RuntimeConfig{
				Variables: map[string]RuntimeVariable{
					"foo": {
						Default:  "bar",
						Required: true,
					},
				},
				KeyValueStore: map[string]KeyValueStore{
					"foo": {
						Type:      RuntimeConfigTypeCosmos,
						Key:       "test-key",
						Account:   "test-account",
						Database:  "test-db",
						Container: "test-container",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadRuntimeConfig(tt.args.runtimeConfigContents)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadRuntimeConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadRuntimeConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
