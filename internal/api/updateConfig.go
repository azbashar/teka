package api

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/A-Bashar/Teka-Finance/internal/config"
)

func updateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		enableCORS(w, r)
		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}

	enableCORS(w, r)

	var partial map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&partial); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	merged := mergeStruct(config.Cfg, partial).(config.Config)
	config.Cfg = merged

	configFile, err := config.GetConfigPath()
	if err != nil {
		http.Error(w, "Failed to get config path: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := config.SaveConfig(configFile); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the full updated config
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config.Cfg)
}

// mergeStruct recursively merges map[string]interface{} into struct
func mergeStruct(orig interface{}, updates map[string]interface{}) interface{} {
	vOrig := reflect.ValueOf(orig)
	if vOrig.Kind() == reflect.Ptr {
		vOrig = vOrig.Elem()
	}
	tOrig := vOrig.Type()

	merged := reflect.New(tOrig).Elem()
	merged.Set(vOrig)

	for i := 0; i < tOrig.NumField(); i++ {
		field := tOrig.Field(i)
		fieldVal := merged.Field(i)
		fieldName := field.Name

		if updateVal, ok := updates[fieldName]; ok {
			switch fieldVal.Kind() {
			case reflect.Struct:
				if m, ok := updateVal.(map[string]interface{}); ok {
					fieldVal.Set(reflect.ValueOf(mergeStruct(fieldVal.Interface(), m)))
				}
			case reflect.Slice:
				if slice, ok := updateVal.([]interface{}); ok {
					sliceVal := reflect.MakeSlice(fieldVal.Type(), len(slice), len(slice))
					for idx, item := range slice {
						if fieldVal.Type().Elem().Kind() == reflect.Struct {
							if m, ok := item.(map[string]interface{}); ok {
								sliceVal.Index(idx).Set(reflect.ValueOf(mergeStruct(reflect.New(fieldVal.Type().Elem()).Elem().Interface(), m)))
							}
						} else {
							sliceVal.Index(idx).Set(reflect.ValueOf(item))
						}
					}
					fieldVal.Set(sliceVal)
				}
			default:
				fieldVal.Set(reflect.ValueOf(updateVal).Convert(fieldVal.Type()))
			}
		}
	}

	return merged.Interface()
}
