
func (m *default{{.upperStartCamelObject}}Model) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", {{.primaryKeyLeft}}, primary)
}

func (m *default{{.upperStartCamelObject}}Model) queryPrimary(conn *gorm.DB, v, primary interface{}) error {
	return conn.Model(&{{.upperStartCamelObject}}{}).Where("{{.originalPrimaryField}} = ?", primary).Take(v).Error
}
