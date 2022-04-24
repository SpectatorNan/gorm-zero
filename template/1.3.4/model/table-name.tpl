
func (m *default{{.upperStartCamelObject}}Model) tableName() string {
	return m.table
}

func ({{.upperStartCamelObject}}) TableName() string {
	model := default{{.upperStartCamelObject}}Model{}
  	return model.tableName()
}