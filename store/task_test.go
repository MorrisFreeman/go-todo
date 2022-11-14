package store

import (
	"context"
	"testing"

	"github.com/MorrisFreeman/go-todo/clock"
	"github.com/MorrisFreeman/go-todo/entity"
	"github.com/MorrisFreeman/go-todo/testutil"
	"github.com/MorrisFreeman/go-todo/testutil/fixture"
	"github.com/google/go-cmp/cmp"
)

func TestRepository_ListTasks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	// entity.Taskを作成する他のテストケースと混ざるとテストに失敗するので
	// トランザクションを貼ることでテーブルの状態を隔離する
	tx, err := testutil.OpenDBForTest(t).BeginTxx(ctx, nil)

	// テストが完了したらロールバック
	t.Cleanup(func() { _ = tx.Rollback() })
	if err != nil {
		t.Fatal(err)
	}
	wanttUserID, wants := prepareTasks(ctx, t, tx)

	sut := &Reposytory{}
	gots, err := sut.ListTasks(ctx, tx, wanttUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d := cmp.Diff(gots, wants); len(d) != 0 {
		t.Errorf("differs: (-got +want)\n%s", d)
	}
}

func prepareUser(ctx context.Context, t *testing.T, db Execer) entity.UserID {
	t.Helper()

	u := fixture.User(nil)
	result, err := db.ExecContext(ctx,
		`INSERT INTO user (name, password, role, created, modified)
		VALUES (?, ?, ?, ?, ?);`,
		u.Name, u.Password, u.Role, u.Created, u.Modified,
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("got user_id: %v", err)
	}

	return entity.UserID(id)
}

func prepareTasks(ctx context.Context, t *testing.T, con Execer) (entity.UserID, entity.Tasks) {
	t.Helper()

	if _, err := con.ExecContext(ctx, "DELETE FROM task;"); err != nil {
		t.Logf("failed to initialize task: %v", err)
	}

	userID := prepareUser(ctx, t, con)
	otherUserID := prepareUser(ctx, t, con)

	c := clock.FixedClocker{}
	wants := entity.Tasks{
		{
			UserID: userID,
			Title:  "want task 1", Status: "todo",
			Created: c.Now(), Modified: c.Now(),
		},
		{
			UserID: userID,
			Title:  "want task 2", Status: "todo",
			Created: c.Now(), Modified: c.Now(),
		},
	}
	tasks := entity.Tasks{
		wants[0],
		{
			UserID: otherUserID,
			Title:  "want task 3", Status: "todo",
			Created: c.Now(), Modified: c.Now(),
		},
		wants[1],
	}

	result, err := con.ExecContext(ctx,
		`INSERT INTO task (user_id, title, status, created, modified)
			VALUES
				(?, ?, ?, ?, ?),
				(?, ?, ?, ?, ?),
				(?, ?, ?, ?, ?);`,
		tasks[0].UserID, tasks[0].Title, tasks[0].Status, tasks[0].Created, tasks[0].Modified,
		tasks[1].UserID, tasks[1].Title, tasks[1].Status, tasks[1].Created, tasks[1].Modified,
		tasks[2].UserID, tasks[2].Title, tasks[2].Status, tasks[2].Created, tasks[2].Modified,
	)
	if err != nil {
		t.Fatal(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	tasks[0].ID = entity.TaskID(id)
	tasks[1].ID = entity.TaskID(id + 1)
	tasks[2].ID = entity.TaskID(id + 2)
	return userID, wants
}
