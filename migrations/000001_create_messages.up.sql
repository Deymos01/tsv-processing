CREATE TABLE IF NOT EXISTS messages (
                                        id                      BIGSERIAL PRIMARY KEY,
                                        number                  INTEGER          NOT NULL,
                                        mqtt                    TEXT             NOT NULL DEFAULT '',
                                        inv_id                  TEXT             NOT NULL DEFAULT '',
                                        unit_guid               TEXT             NOT NULL,
                                        message_id              TEXT             NOT NULL DEFAULT '',
                                        message_text            TEXT             NOT NULL DEFAULT '',
                                        context                 TEXT             NOT NULL DEFAULT '',
                                        message_class           TEXT             NOT NULL DEFAULT '',
                                        message_level           TEXT             NOT NULL DEFAULT '',
                                        variable_zone           TEXT             NOT NULL DEFAULT '',
                                        variable_address        TEXT             NOT NULL DEFAULT '',
                                        use_as_block_start      BOOLEAN          NOT NULL DEFAULT FALSE,
                                        type                    TEXT             NOT NULL DEFAULT '',
                                        bit_number_in_register  INTEGER          NOT NULL DEFAULT 0,
                                        invert_bit              BOOLEAN          NOT NULL DEFAULT FALSE,
                                        source_file             TEXT             NOT NULL,
                                        created_at              TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_unit_guid ON messages (unit_guid);
CREATE INDEX idx_messages_source_file ON messages (source_file);