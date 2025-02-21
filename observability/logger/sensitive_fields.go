package oblogger

import (
	"encoding/json"
	"reflect"
	"regexp"

	"github.com/gianglt2198/platforms/pkg/utils"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type (
	SensitiveFieldEncoder struct {
		zapcore.Encoder
		cfg             zapcore.EncoderConfig
		sensitiveFields map[string]string
	}
)

func NewSensitiveFieldsEncoder(config zapcore.EncoderConfig, sensitiveFields map[string]string) zapcore.Encoder {
	encoder := zapcore.NewJSONEncoder(config)
	return &SensitiveFieldEncoder{encoder, config, sensitiveFields}
}

func (e *SensitiveFieldEncoder) EncodeEntry(
	entry zapcore.Entry,
	fields []zapcore.Field,
) (*buffer.Buffer, error) {
	filtered := make([]zapcore.Field, 0, len(fields))

	for _, field := range fields {
		field = e.MaskJSON(field)
		filtered = append(filtered, field)
	}

	return e.Encoder.EncodeEntry(entry, filtered)
}

func (e *SensitiveFieldEncoder) Clone() zapcore.Encoder {
	return &SensitiveFieldEncoder{
		Encoder: e.Encoder.Clone(),
	}
}

func (e *SensitiveFieldEncoder) MaskJSON(field zapcore.Field) zapcore.Field {
	switch field.Type {
	case zapcore.StringType:
		field.String = regexp.MustCompile(".").ReplaceAllLiteralString(field.String, "*")
	case zapcore.ReflectType:
		jsonMap := make(map[string]interface{})
		if err := utils.TransformMaptoStruct(field.Interface, &jsonMap); err != nil {
			return field
		}
		jsonMap = e.maskMap(jsonMap)
		if jsonBytes, err := json.Marshal(&jsonMap); err == nil {
			field.Interface = string(jsonBytes)
			field.String = string(jsonBytes)
		}
	}

	return field
}

func (e *SensitiveFieldEncoder) maskMap(jsonMap map[string]interface{}) map[string]interface{} {
	for key, value := range jsonMap {
		if value == nil {
			continue
		}

		if reflect.TypeOf(value).Kind() == reflect.Map {
			temporaryMap := value.(map[string]interface{})
			e.maskMap(temporaryMap)
		} else {
			if sv, ok := e.sensitiveFields[key]; ok {
				if reflect.TypeOf(value).Kind() == reflect.String {
					switch sv {
					case "":
						jsonMap[key] = value
					case "MASKALL":
						jsonMap[key] = regexp.MustCompile(".").ReplaceAllLiteralString(reflect.ValueOf(value).String(), "*")
					default:
						re := regexp.MustCompile(sv)
						reValues := re.FindStringSubmatch(reflect.ValueOf(value).String())
						if reValues != nil {
							reNames := re.SubexpNames()
							var maskValue string
							for i := 1; i < len(reNames); i++ {
								if reNames[i] == "MASK" {
									maskValue += regexp.MustCompile(".").ReplaceAllLiteralString(reValues[i], "*")
								} else {
									maskValue += reValues[i]
								}
							}
							jsonMap[key] = maskValue
						} else {
							jsonMap[key] = value
						}
					}
				}
			}
		}
	}
	return jsonMap
}
