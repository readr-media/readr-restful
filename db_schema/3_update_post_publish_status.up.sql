-- OLD ACTIVE
-- 1: active（顯示在前台）, 0: deactive（僅存在資料庫）, 2: pending（已提交待審核）,3: draft（草稿）,4:unpubilsh（已提交但僅顯示在後台）
-- NEW PUBLISH_STATUS
-- 0: unpublish（僅顯示在後台）, 1: draft（草稿）, 2: publish（已發布）,3: schedule（排程）,4:pending（等待審核）
-- NEW ACTIVE
-- 0: deactive（僅存在資料庫）, 1: active（顯示在前台
ALTER TABLE posts ADD COLUMN publish_status tinyint DEFAULT 0;
UPDATE posts SET publish_status=0 WHERE active=0;
UPDATE posts SET publish_status=2 WHERE active=1;
UPDATE posts SET publish_status=4, active=1 WHERE active=2;
UPDATE posts SET publish_status=1, active=1 WHERE active=3;
UPDATE posts SET publish_status=0, active=1 WHERE active=4;