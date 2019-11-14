package cards

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/internal/rrsql"
)

func (a *newscardAPI) buildGetStmt(args *NewsCardArgs) (query string, values []interface{}) {
	selectedFields := []string{"newscards.*"}
	var restricts string

	restricts, restrictVals := args.parse()
	resultLimit, resultLimitVals := args.parseResultLimit()
	values = append(values, restrictVals...)
	values = append(values, resultLimitVals...)

	query = fmt.Sprintf(`
		SELECT %s FROM newscards %s `,
		strings.Join(selectedFields, ","),
		restricts+resultLimit,
	)

	return query, values
}

func BuildSyncStmts(postID uint32, cards []NewsCard) (stmts []*rrsql.PipelineStmt, err error) {

	var cardsToBeUpdated, cardsToBeInserted []NewsCard
	for _, card := range cards {
		if card.ID == 0 {
			cardsToBeInserted = append(cardsToBeInserted, card)
		} else {
			cardsToBeUpdated = append(cardsToBeUpdated, card)
		}
	}

	if postID != 0 {
		var cardIDs []uint32
		for _, card := range cardsToBeUpdated {
			cardIDs = append(cardIDs, card.ID)
		}
		stmt, err := buildDeleteStmt(postID, cardIDs)
		if err != nil {
			return stmts, err
		}
		stmts = append(stmts, stmt)
	}
	if len(cardsToBeUpdated) > 0 && postID != 0 {
		for _, card := range cardsToBeUpdated {
			stmts = append(stmts, buildUpdateStmt(card))
		}
	}

	if len(cardsToBeInserted) > 0 {
		for _, card := range cardsToBeInserted {
			stmts = append(stmts, buildInsertStmt(postID, card))
		}
	}

	return stmts, nil
}

func buildDeleteStmt(postID uint32, cardIDs []uint32) (stmt *rrsql.PipelineStmt, err error) {
	var query string
	var args []interface{}
	if len(cardIDs) == 0 {
		query = fmt.Sprintf("UPDATE newscards SET active = %d WHERE post_id = ?", config.Config.Models.Cards["deactive"])
		args = []interface{}{postID}
	} else {
		query = fmt.Sprintf("UPDATE newscards SET active = %d WHERE post_id = ? AND id NOT in (?)", config.Config.Models.Cards["deactive"])
		args = []interface{}{postID, cardIDs}
		query, args, err = sqlx.In(query, args...)
		if err != nil {
			return nil, err
		}
		query = rrsql.DB.Rebind(query)
	}
	return &rrsql.PipelineStmt{Query: query, Args: args}, nil
}

func buildUpdateStmt(card NewsCard) *rrsql.PipelineStmt {
	tags := rrsql.GetStructDBTags("partial", card)
	fields := rrsql.MakeFieldString("update", `%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE newscards SET %s WHERE id = :id`,
		strings.Join(fields, ", "))

	return &rrsql.PipelineStmt{
		Query:     query,
		NamedArgs: card,
		NamedExec: true,
	}
}

func buildInsertStmt(postID uint32, card NewsCard) *rrsql.PipelineStmt {
	postIDString := config.Config.SQL.TrasactionIDPlaceholder
	if postID != 0 {
		postIDString = strconv.Itoa(int(postID))
	}

	tags := rrsql.GetStructDBTags("partial", card)
	query := fmt.Sprintf(`INSERT INTO newscards (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))

	query = strings.Replace(query, ":post_id", postIDString, -1)

	return &rrsql.PipelineStmt{
		Query:        query,
		NamedArgs:    card,
		NamedExec:    true,
		RowsAffected: true,
	}
}
