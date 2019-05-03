//Cache

package cache

import (
	"errors"
	mapstruct "github.com/ottemo/mapstructure"
	"github.com/jxbdlut/cache/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"reflect"
	"strconv"
	"time"
	"github.com/jinzhu/gorm"
)

var log *zap.SugaredLogger
var cachePrefix = []byte("")

var (
	ErrKeyNotExist = errors.New("keys not exists")
)

func SetCachePrefix(str string) {
	cachePrefix = []byte(str)
}

func GetCachePrefix() []byte {
	return cachePrefix
}

type CacheRedis interface {
	Set(key string, b interface{}) error
	Expire(key string, expiration time.Duration) (bool, error)
	Get(key string) ([]byte, error)
	HGetAll(key string) (interface{}, error)
	HMSet(key string, fields map[string]interface{}) (string, error)
	HSet(key string, field string, v interface{}) (bool, error)
	Keys(key string) ([]string, error)
	Del(key string) (int64, error)
	Exists(key string) (bool, error)
	Close() error
}

type Module interface {
	TableName() string
	GetKey() interface{}
}

type Cache struct {
	CacheRedis
	DataBase
}

func NewCache() *Cache {
	log = logging.Logger(zapcore.DebugLevel)
	return &Cache{}
}

func IsLackField(obj interface{}, mapValue map[string]string) bool {
	t := reflect.TypeOf(obj).Elem()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if nocache := field.Tag.Get("no_cache"); len(nocache) != 0 {
			continue
		}
		if _, exist:= mapValue[field.Name]; !exist {
			return true
		}
	}
	return false
}

func Struct2Map(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj).Elem()
	v := reflect.ValueOf(obj).Elem()

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if nocache := field.Tag.Get("no_cache"); len(nocache) != 0 {
			continue
		}
		switch v.Field(i).Interface().(type) {
		case time.Time:
			data[t.Field(i).Name] = v.Field(i).Interface().(time.Time).Format(time.RFC3339)
		default:
			data[t.Field(i).Name] = v.Field(i).Interface()
		}
	}
	return data
}

func myDecoder(val *reflect.Value, data interface{}) (interface{}, error) {
	if val.Type().String() == "time.Time" {
		value, err := time.Parse(time.RFC3339, data.(string))
		val.Set(reflect.ValueOf(value))
		return nil, err
	}
	return data, nil
}

func getDecoder(result interface{}) (*mapstruct.Decoder, error) {
	return mapstruct.NewDecoder(&mapstruct.DecoderConfig{
		AdvancedDecodeHook: myDecoder,
		TagName:            "",
		Result:             result,
		WeaklyTypedInput:   true})
}

func (c *Cache) SetDebug(b bool) {
	if b {
		log = logging.Logger(zapcore.DebugLevel)
	} else {
		log = logging.Logger(zapcore.ErrorLevel)
	}
}

func (c *Cache) SetCacheDatabase(name string, db *gorm.DB) {
	db.Callback().Update().After("gorm:update").Register("cache:after_update", c.AfterUpdate)
	c.SetDatabase(name, db)
}

func (c *Cache) AfterUpdate(scope *gorm.Scope) {
	value := scope.Value
	reflect.TypeOf(value)
	c.CacheRedis.Del(c.CacheKey(value.(Module)))
}

func (c *Cache) Get(mode Module) error {
	needReadDb := false
	key := c.CacheKey(mode)
	log.Debugf("key:%v mode:%v", key, mode)
	exist, err := c.CacheRedis.Exists(key)
	log.Debugf("exist:%v err:%v", exist, err)
	var result interface{}
	if err == nil && exist {
		result, err = c.CacheRedis.HGetAll(key)
		if err != nil {
			log.Errorf("redis HGetAll err:%v", err)
			return err
		}
		//检查缓存中是否缺少字段
		needReadDb = IsLackField(mode, result.(map[string]string))
	} else {
		needReadDb = true
	}
	if needReadDb {
		log.Debugf("cache Keys key:%v err:%v", key, err)
		if err = c.GetReadDb().First(mode, mode.GetKey()).Error; err != nil {
			log.Errorf("db first err:%v key:%v", err, mode.GetKey())
			return err
		} else {
			m := Struct2Map(mode)
			c.CacheRedis.HMSet(key, m)
			c.CacheRedis.Expire(key, 7 * 24 * 3600 * time.Second)
		}
	} else {
		decode, err := getDecoder(mode)
		if err != nil {
			log.Errorf("mapstruct getDecoder err:%v", err)
			return err
		}
		err = decode.Decode(result)
		if err != nil {
			log.Errorf("Decode err:%v", err)
			return err
		}
		log.Debugf("result:%v mode:%v", result, mode)
	}
	return nil
}

func (c *Cache) CacheKey(mode Module) string {
	c.CacheRedis = c.GetCacheClient()
	value := reflect.ValueOf(mode).Elem()
	typeOf := reflect.TypeOf(mode).Elem()
	str := cachePrefix

	for i := 0; i < value.NumField(); i++ {
		field := typeOf.Field(i)
		if name := field.Tag.Get("cache"); len(name) > 0 {
			val := value.Field(i)
			if i != 0 {
				str = append(str, []byte(":")...)
			}
			str = append(str, c.fieldToByte(val.Interface())...)
			str = append(str, []byte(":"+name)...)

		}
	}
	return string(str)
}

func (c *Cache) fieldToByte(value interface{}) (str []byte) {
	var st = []byte("*")
	typ := reflect.TypeOf(value)
	val := reflect.ValueOf(value)
	switch typ.Kind() {
	case reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uint8, reflect.Uint16:
		if val.Uint() <= 0 {
			str = append(str, st...)
		} else {
			str = strconv.AppendUint(str, val.Uint(), 10)
		}
	case reflect.Int32, reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16:
		if val.Int() <= 0 {
			str = append(str, st...)
		} else {
			str = strconv.AppendInt(str, val.Int(), 10)
		}
	case reflect.Float32, reflect.Float64:
		if val.Float() == 0.0 {
			str = append(str, st...)
		} else {
			str = strconv.AppendFloat(str, val.Float(), 'f', 0, 64)
		}
	case reflect.String:
		if val.Len() <= 0 {
			str = append(str, st...)
		} else {
			str = append(str, []byte(val.String())...)
		}
	case reflect.Bool:
		switch val.Bool() {
		case true:
			str = append(str, []byte("true")...)
		case false:
			str = append(str, []byte("false")...)
		}
	default:
		switch value.(type) {
		case time.Time:
			str = append(str, []byte(value.(time.Time).Format(time.RFC1123Z))...)
		}
	}
	return
}
