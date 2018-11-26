package poll

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/readr-media/readr-restful/models"
)

// Poll is the struct mapping to table polls
type Poll struct {
	ID          int64             `json:"id" db:"id"`
	Status      int64             `json:"status" db:"status"`
	Active      int64             `json:"active" db:"active"`
	Title       models.NullString `json:"title" db:"title"`
	Description models.NullString `json:"description" db:"description"`
	TotalVote   int64             `json:"total_vote" db:"total_vote"`
	Frequency   models.NullString `json:"frequency" db:"frequency"`
	StartAt     models.NullTime   `json:"start_at" db:"start_at"`
	EndAt       models.NullTime   `json:"end_at" db:"end_at"`
	MaxChoice   int64             `json:"max_choice" db:"max_choice"`
	Changeable  int64             `json:"changeable" db:"changeable"`
	PublishedAt models.NullInt    `json:"published_at" db:"published_at"`
	CreatedAt   models.NullTime   `json:"created_at" db:"created_at"`
	CreatedBy   models.NullInt    `json:"created_by" db:"created_by"`
	UpdatedAt   models.NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy   models.NullInt    `json:"updated_by" db:"updated_by"`
}

// Choice is the struct mapping to table polls_choices
// storing choice data for each poll
type Choice struct {
	ID         int64             `json:"id" db:"id"`
	Choice     models.NullString `json:"choice" db:"choice"`
	TotalVote  models.NullInt    `json:"total_vote" db:"total_vote"`
	PollID     models.NullInt    `json:"poll_id" db:"poll_id"`
	Active     models.NullInt    `json:"active" db:"active"`
	GroupOrder models.NullInt    `json:"group_order" db:"group_order"`
	CreatedAt  models.NullTime   `json:"created_at" db:"created_at"`
	CreatedBy  models.NullInt    `json:"created_by" db:"created_by"`
	UpdatedAt  models.NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy  models.NullInt    `json:"updated_by" db:"updated_by"`
}

// ChosenChoice is the mapping struct for table polls_chosen choice
// to record the choosing history for every users
type ChosenChoice struct {
	ID        int64           `json:"id" db:"id"`
	MemberID  int64           `json:"member_id" db:"member_id"`
	PollID    int64           `json:"poll_id" db:"poll_id"`
	ChoiceID  int64           `json:"choice_id" db:"choice_id"`
	CreatedAt models.NullTime `json:"created_at" db:"created_at"`
}

// ChoicesEmbeddedPoll is a single complete poll struct
// Corresponding choices are embedded in the return values
type ChoicesEmbeddedPoll struct {
	Poll
	Choices []Choice `json:"choices,omitempty" db:"choices"`
}

type pollInterface interface {
	Get(filters *GetPollsFilter) (polls []ChoicesEmbeddedPoll, err error)
	Insert(p ChoicesEmbeddedPoll) (err error)
	Update(poll Poll) (err error)
}

type pollData struct{}

type choiceInterface interface {
	Get(id int) (choices []Choice, err error)
	Insert(choices []Choice) (err error)
	Update(choices []Choice) (err error)
}

type choiceData struct{}

// GetStructTags is designed to be a public function aiming at future reuse.
// It is also designed to be a variadic function.
// The first argument is skip fields, which denotes the field we don't want
func GetStructTags(mode string, tagname string, input interface{}, options ...interface{}) []string {

	columns := make([]string, 0)
	value := reflect.ValueOf(input)

	// Originally used to rule out id field when insert
	var skipFields []string
	if options != nil {
		skipFields = options[0].([]string)
	}

FindTags:
	for i := 0; i < value.NumField(); i++ {

		field := value.Type().Field(i)
		fieldType := field.Type
		fieldValue := value.Field(i)

		// Use Type() to get struct tags
		tag := value.Type().Field(i).Tag.Get(tagname)
		// Skip fields if there are denoted
		if len(skipFields) > 0 {

			for _, f := range skipFields {
				if tag == f {
					fmt.Printf("Found skip fields %s!\n", f)
					continue FindTags
				}
			}
		}

		if mode == "full" {
			columns = append(columns, tag)
		} else if mode == "non-null" {
			// Append each tag for non-null field
			switch fieldType.Name() {
			case "string":
				if fieldValue.String() != "" {
					columns = append(columns, tag)
				}
			case "int64", "int":
				if fieldValue.Int() != 0 {
					columns = append(columns, tag)
				}
			case "uint32":
				if fieldValue.Uint() != 0 {
					columns = append(columns, tag)
				}
			case "NullString", "NullInt", "NullTime", "NullBool":
				if fieldValue.FieldByName("Valid").Bool() {
					columns = append(columns, tag)
				}
			default:
				fmt.Println("unrecognised format: ", value.Field(i).Type())
			}
		}
	}
	return columns
}

type sqlsv struct {
	statement string
	variable  interface{}
}

type sqlfield struct {
	table   string
	pattern string
	fields  []string
}

func (sf sqlfield) formatter() string {
	var results []string
FieldLoop:
	for _, f := range sf.fields {
		if f == "id" {
			results = append(results, fmt.Sprintf(`IFNULL(%s.id, 0) "%s.id"`, sf.table, sf.table))
			continue FieldLoop
		}
		switch sf.pattern {
		case `%s.%s "%s.%s"`:
			results = append(results, fmt.Sprintf(sf.pattern, sf.table, f, sf.table, f))
		case `%s.%s`:
			results = append(results, fmt.Sprintf(sf.pattern, sf.table, f))
		default:
			fmt.Printf("could not parse:%s\n", sf.pattern)
		}
	}
	return strings.Join(results, ", ")
}

// SQLO , not SOLO, stands for "SQL Object".
// It tries to mapping SQL statement to struct
type SQLO struct {

	// table hosts the table name for select
	table string

	// fields hosts all the fields that will appear in SELECT statement
	fields []sqlfield

	// join provide table to be joined, in string form
	join []string

	// where comprises SQL statements strings in 'WHERE' section
	where []sqlsv

	// order by maps to the ORDER BY section in SQL
	orderby string

	// pagination is the limit string, in this pattern: LIMIT [max_results] OFFSET [page-1]
	pagination string

	// args comprises all the argument corresponding to placeholders in SQL statements
	// They will be passed into sqlx functions
	args []interface{}
}

func (s *SQLO) FormatOrderBy(orderby string) {

	tmp := strings.Split(orderby, ",")
	for i, v := range tmp {
		if v := strings.TrimSpace(v); strings.HasPrefix(v, "-") {
			tmp[i] = v[1:] + " DESC"
		} else {
			tmp[i] = v
		}
	}
	s.orderby = fmt.Sprintf(" ORDER BY %s", strings.Join(tmp, ", "))
}

func (s *SQLO) GenFields() string {

	var results []string
	for _, field := range s.fields {
		results = append(results, field.formatter())
	}
	return strings.Join(results, ", ")
}

func (s *SQLO) SQL() string {

	base := bytes.NewBufferString(fmt.Sprintf("SELECT %s FROM %s", s.GenFields(), s.table))
	for _, join := range s.join {
		base.WriteString(join)
	}
	if len(s.where) > 0 {
		base.WriteString(" WHERE")
	}
	for _, condition := range s.where {
		base.WriteString(condition.statement)
		s.args = append(s.args, condition.variable)
	}
	if s.orderby != "" {
		base.WriteString(s.orderby)
	}
	if s.pagination != "" {
		base.WriteString(s.pagination)
	}
	base.WriteString(";")
	return base.String()
}

func NewSQLO(options ...func(*SQLO)) *SQLO {
	so := SQLO{}
	for _, option := range options {
		option(&so)
	}
	return &so
}

func (p *pollData) Get(filter *GetPollsFilter) (polls []ChoicesEmbeddedPoll, err error) {

	osql := NewSQLO(func(s *SQLO) {
		s.table = "polls"
		s.fields = append(s.fields, sqlfield{table: "polls", pattern: `%s.%s`, fields: []string{"*"}})
	}, filter.Parse())
	rows, err := models.DB.Queryx(osql.SQL(), osql.args...)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
ScanLoop:
	for rows.Next() {
		// Corresponding struct for joined table
		var poll struct {
			Poll
			Choice `db:"choice"`
		}
		if err = rows.StructScan(&poll); err != nil {
			log.Fatal("Error scan polls\n", err)
			return nil, err
		}
		// Append choices if there is already a poll entry for this poll id
		for i, v := range polls {
			if v.ID == poll.Poll.ID {
				if poll.Choice.ID != 0 {
					polls[i].Choices = append(polls[i].Choices, poll.Choice)
				}
				continue ScanLoop
			}
		}
		// Poll id not existing. Create new poll for this id
		if poll.Choice.ID != 0 {
			polls = append(polls, ChoicesEmbeddedPoll{Poll: poll.Poll, Choices: []Choice{poll.Choice}})
		} else {
			polls = append(polls, ChoicesEmbeddedPoll{Poll: poll.Poll})
		}
	}
	return polls, err
}

// Insert allows consumer to insert a poll at a time.
// This poll could have attached choices, which will be also inserted as well.
// Insert does not allow empty poll with choices, you have to insert poll first.
// If it's needed to insert new choice, use choice api instead.
func (p *pollData) Insert(poll ChoicesEmbeddedPoll) (err error) {

	pollTags := GetStructTags("full", "db", Poll{})
	tx, err := models.DB.Beginx()
	if err != nil {
		log.Printf("Fail to get sql connection: %v\n", err)
		return err
	}
	// Either rollback or commit transaction
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// Insert poll first
	// Notice colon in 'VALUES (:%s)', because strings.Join will not create the first colon
	pollQ := fmt.Sprintf(`INSERT INTO polls (%s) VALUES (:%s)`,
		strings.Join(pollTags, ","), strings.Join(pollTags, ",:"))
	pollInserted, err := tx.NamedExec(pollQ, poll.Poll)
	if err != nil {
		return err
	}
	pollID, err := pollInserted.LastInsertId()
	if err != nil {
		return err
	}
	if len(poll.Choices) > 0 {
		choiceTags := GetStructTags("full", "db", Choice{})

		choiceQ := fmt.Sprintf(`INSERT INTO polls_choices (%s) VALUES (:%s)`,
			strings.Join(choiceTags, ","), strings.Join(choiceTags, ",:"))

		// Change poll_id in options to the poll id we just inserted
		// Batch insert for tx is merged but not released yet, use for loop
		// Info: https://github.com/jmoiron/sqlx/pull/285
		for _, choice := range poll.Choices {
			choice.PollID.Int = pollID
			choice.PollID.Valid = true
			if _, err := tx.NamedExec(choiceQ, choice); err != nil {
				return err
			}
		}
	}
	return nil
}

// Update single row of poll
func (p *pollData) Update(poll Poll) (err error) {

	pollTags := GetStructTags("non-null", "db", poll)
	pollFields := func(tags []string) string {
		var temp []string
		for _, tag := range tags {
			temp = append(temp, fmt.Sprintf(`%s = :%s`, tag, tag))
		}
		return strings.Join(temp, ", ")
	}(pollTags)
	query := fmt.Sprintf(`Update polls SET %s WHERE id = :id`, pollFields)
	if _, err := models.DB.NamedExec(query, poll); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (c *choiceData) Get(pollID int) (results []Choice, err error) {

	Q := `SELECT * FROM polls_choices WHERE poll_id = ?;`
	rows, err := models.DB.Queryx(Q, pollID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var choice Choice
		if err := rows.StructScan(&choice); err != nil {
			return nil, err
		}
		results = append(results, choice)
	}
	return results, err
}

func RepeatString(target string, times int, delimiter string) (result string) {
	var inter []string
	for i := 0; i < times; i++ {
		inter = append(inter, target)
	}
	return strings.Join(inter, delimiter)
}

func (c *choiceData) Insert(choices []Choice) (err error) {

	choiceTags := GetStructTags("full", "db", Choice{})

	choiceQ := fmt.Sprintf(`INSERT INTO polls_choices (%s) VALUES (:%s)`,
		strings.Join(choiceTags, ", "), strings.Join(choiceTags, ", :"))

	tx, err := models.DB.Beginx()
	if err != nil {
		log.Printf("Fail to get sql connection: %v\n", err)
		return err
	}
	// Either rollback or commit transaction
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	// Change poll_id in options to the poll id we just inserted
	// Batch insert for tx is merged but not released yet, use for loop
	// Info: https://github.com/jmoiron/sqlx/pull/285
	for _, choice := range choices {
		// Maybe it could be implemented using Prepare and exec
		// to avoid prepare statement for each loop
		if _, err := tx.NamedExec(choiceQ, choice); err != nil {
			return err
		}
	}
	return nil
}

// Update single choice
func (c *choiceData) Update(choices []Choice) (err error) {

	tx, err := models.DB.Beginx()
	if err != nil {
		log.Printf("Fail to get sql connection: %v\n", err)
		return err
	}
	// Either rollback or commit transaction
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	for _, choice := range choices {

		choiceTags := GetStructTags("non-null", "db", choice)
		choiceFields := func(tags []string) string {
			var temp []string
			for _, tag := range tags {
				temp = append(temp, fmt.Sprintf(`%s = :%s`, tag, tag))
			}
			return strings.Join(temp, ", ")
		}(choiceTags)
		query := fmt.Sprintf(`Update polls_choices SET %s WHERE id = :id`, choiceFields)
		if _, err = tx.NamedExec(query, choice); err != nil {
			log.Fatal(err)
			return err
		}
	}
	return nil
}

// PollData is the pointer instance of pollData struct, which implements pollInterface
// It provides data layer abstraction
var PollData = new(pollData)

// ChoiceData is the pointer instance of choiceData struct, which implements choiceInterface,
// to provide choice database abstraction
var ChoiceData = new(choiceData)
