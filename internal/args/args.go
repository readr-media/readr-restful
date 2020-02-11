package args

type ArgsParser interface {
	//ParseQuery() (restricts string, values []interface{})
	ParseCountQuery() (restricts string, values []interface{})
}
