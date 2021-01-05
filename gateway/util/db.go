package util

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

func GetHashByTag(ctx context.Context, dbAddr string, tags ...string) []string {
	back := []string{}
	conn, err := pgx.Connect(ctx, dbAddr)
	if err != nil {
		return back
	}
	defer conn.Close(ctx)

	if len(tags) == 0 {
		return back
	}

	sql := fmt.Sprintf(`SELECT hash FROM %s WHERE `, TBTag)
	for i, tag := range tags {
		if i == 0 {
			sql = sql + fmt.Sprintf("tag='%s' ", tag)
			continue
		}
		sql = sql + fmt.Sprintf("AND tag='%s' ", tag)
	}

	Log().Debugf("select hash with sql:%s", sql)
	rows, err := conn.Query(ctx, sql)
	if err != nil {
		Log().Errorf("select hash with sql:%s fail, err:%v", sql, err)
		return back
	}
	for rows.Next() {
		h := ""
		if err := rows.Scan(&h); err != nil {
			Log().Errorf("scan sql hash fail, err:%v", err)
			continue
		}
		back = append(back, h)
	}
	return back
}

func AddHashTag(ctx context.Context, dbAddr, hash, tag string) error {
	conn, err := pgx.Connect(ctx, dbAddr)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	sql := fmt.Sprintf(`INSERT INTO %s(hash, tag) VALUES($1, $2) `, TBTag)
	_, err = conn.Exec(ctx, sql, hash, tag)
	return err
}
