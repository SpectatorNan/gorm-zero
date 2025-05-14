{{if .withCache}}
func (m *default{{.upperStartCamelObject}}Model) GetCacheKeys(data *{{.upperStartCamelObject}}) []string {
    if data == nil {
        return []string{}
    }
    {{.keys}}
    cacheKeys := []string{
        {{.keyValues}},
    }
    cacheKeys = append(cacheKeys, m.customCacheKeys(data)...)
    return cacheKeys
}
{{end}}

func (m *default{{.upperStartCamelObject}}Model) Insert(ctx context.Context, tx *gorm.DB, data *{{.upperStartCamelObject}}) error {
	{{if .withCache}}
    err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
        if tx != nil {
            db = tx
        }
        return db.Create(&data).Error
	}, m.GetCacheKeys(data)...){{else}}db := m.conn
        if tx != nil {
            db = tx
        }
        err := db.WithContext(ctx).Create(&data).Error{{end}}
	return err
}
func (m *default{{.upperStartCamelObject}}Model) BatchInsert(ctx context.Context, tx *gorm.DB, news []{{.upperStartCamelObject}}) error {
	{{if .withCache}}
    err := batchx.BatchExecCtxV2(ctx, m, news, func(conn *gorm.DB) error {
    db := conn
    		for _, v := range news {
    			if err := db.Create(&v).Error; err != nil {
    				return err
    			}
    		}
    		return nil
    	},tx){{else}}db := m.conn
        if tx != nil {
            db = tx
        }
        err := db.Create(&news).Error{{end}}
	return err
}
