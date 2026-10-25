CREATE TABLE metadata_settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    credential_type TEXT NOT NULL,
    language TEXT NOT NULL,
    credential TEXT NOT NULL
);
INSERT INTO metadata_settings VALUES (1, 'bearer', 'zh-CN', '');
