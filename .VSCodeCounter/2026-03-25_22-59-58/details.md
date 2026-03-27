# Details

Date : 2026-03-25 22:59:58

Directory /home/kirill/Yandex.Disk/myProjects/go/ya-practicum/memotogo

Total : 58 files,  4259 codes, 38 comments, 519 blanks, all 4816 lines

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)

## Files
| filename | language | code | comment | blank | total |
| :--- | :--- | ---: | ---: | ---: | ---: |
| [cmd/bot/main.go](/cmd/bot/main.go) | Go | 116 | 0 | 20 | 136 |
| [deploy/docker-compose\_psql.yaml](/deploy/docker-compose_psql.yaml) | YAML | 16 | 1 | 0 | 17 |
| [go.mod](/go.mod) | Go Module File | 50 | 0 | 5 | 55 |
| [go.sum](/go.sum) | Go Checksum File | 1,102 | 0 | 1 | 1,103 |
| [internal/audio/audio.go](/internal/audio/audio.go) | Go | 66 | 0 | 19 | 85 |
| [internal/client/gigachat/gigachat.go](/internal/client/gigachat/gigachat.go) | Go | 71 | 0 | 18 | 89 |
| [internal/client/gigachat/type.go](/internal/client/gigachat/type.go) | Go | 74 | 1 | 14 | 89 |
| [internal/client/http.go](/internal/client/http.go) | Go | 18 | 0 | 3 | 21 |
| [internal/client/oauth/oauth.go](/internal/client/oauth/oauth.go) | Go | 76 | 0 | 8 | 84 |
| [internal/client/oauth/type.go](/internal/client/oauth/type.go) | Go | 23 | 0 | 6 | 29 |
| [internal/client/salute/adapter.go](/internal/client/salute/adapter.go) | Go | 1 | 0 | 2 | 3 |
| [internal/client/salute/async\_transcriptor.go](/internal/client/salute/async_transcriptor.go) | Go | 239 | 1 | 28 | 268 |
| [internal/client/salute/async\_types.go](/internal/client/salute/async_types.go) | Go | 107 | 0 | 21 | 128 |
| [internal/client/salute/salute.go](/internal/client/salute/salute.go) | Go | 25 | 0 | 5 | 30 |
| [internal/client/salute/sync\_transcriptor.go](/internal/client/salute/sync_transcriptor.go) | Go | 54 | 0 | 8 | 62 |
| [internal/client/salute/sync\_types.go](/internal/client/salute/sync_types.go) | Go | 19 | 0 | 4 | 23 |
| [internal/client/telegram.go](/internal/client/telegram.go) | Go | 17 | 0 | 5 | 22 |
| [internal/config/config.go](/internal/config/config.go) | Go | 125 | 2 | 26 | 153 |
| [internal/deps/deps.go](/internal/deps/deps.go) | Go | 13 | 0 | 5 | 18 |
| [internal/handler/handler.go](/internal/handler/handler.go) | Go | 47 | 0 | 7 | 54 |
| [internal/handler/onaudio.go](/internal/handler/onaudio.go) | Go | 82 | 0 | 6 | 88 |
| [internal/handler/onfind.go](/internal/handler/onfind.go) | Go | 54 | 1 | 6 | 61 |
| [internal/handler/onget.go](/internal/handler/onget.go) | Go | 65 | 1 | 6 | 72 |
| [internal/handler/onlist.go](/internal/handler/onlist.go) | Go | 58 | 2 | 9 | 69 |
| [internal/handler/onvoice.go](/internal/handler/onvoice.go) | Go | 67 | 0 | 6 | 73 |
| [internal/handler/start.go](/internal/handler/start.go) | Go | 67 | 0 | 6 | 73 |
| [internal/llm/adapter/gigachat.go](/internal/llm/adapter/gigachat.go) | Go | 100 | 0 | 21 | 121 |
| [internal/llm/interface.go](/internal/llm/interface.go) | Go | 8 | 0 | 5 | 13 |
| [internal/llm/llm.go](/internal/llm/llm.go) | Go | 100 | 0 | 14 | 114 |
| [internal/llm/llmtype/llmtype.go](/internal/llm/llmtype/llmtype.go) | Go | 57 | 1 | 13 | 71 |
| [internal/llm/prompts.go](/internal/llm/prompts.go) | Go | 33 | 0 | 15 | 48 |
| [internal/llm/state.go](/internal/llm/state.go) | Go | 54 | 0 | 10 | 64 |
| [internal/llm/tool/extractor.go](/internal/llm/tool/extractor.go) | Go | 12 | 0 | 4 | 16 |
| [internal/llm/tool/get\_meeting.go](/internal/llm/tool/get_meeting.go) | Go | 103 | 0 | 15 | 118 |
| [internal/llm/tool/list\_meetings.go](/internal/llm/tool/list_meetings.go) | Go | 93 | 0 | 15 | 108 |
| [internal/llm/tool/registry.go](/internal/llm/tool/registry.go) | Go | 14 | 0 | 6 | 20 |
| [internal/middleware/middleware.go](/internal/middleware/middleware.go) | Go | 50 | 0 | 6 | 56 |
| [internal/model/meeting.go](/internal/model/meeting.go) | Go | 37 | 0 | 6 | 43 |
| [internal/model/transcription.go](/internal/model/transcription.go) | Go | 28 | 0 | 5 | 33 |
| [internal/model/user.go](/internal/model/user.go) | Go | 10 | 0 | 3 | 13 |
| [internal/queue/queue.go](/internal/queue/queue.go) | Go | 16 | 0 | 4 | 20 |
| [internal/repository/generic.go](/internal/repository/generic.go) | Go | 109 | 3 | 17 | 129 |
| [internal/repository/pgstorage.go](/internal/repository/pgstorage.go) | Go | 105 | 0 | 13 | 118 |
| [internal/repository/transcriptions.go](/internal/repository/transcriptions.go) | Go | 185 | 0 | 26 | 211 |
| [internal/repository/type.go](/internal/repository/type.go) | Go | 38 | 0 | 7 | 45 |
| [internal/utils/message.go](/internal/utils/message.go) | Go | 15 | 0 | 4 | 19 |
| [internal/workers/creator.go](/internal/workers/creator.go) | Go | 71 | 0 | 10 | 81 |
| [internal/workers/executor.go](/internal/workers/executor.go) | Go | 121 | 0 | 13 | 134 |
| [internal/workers/interfaces.go](/internal/workers/interfaces.go) | Go | 15 | 0 | 5 | 20 |
| [internal/workers/notifier.go](/internal/workers/notifier.go) | Go | 69 | 0 | 8 | 77 |
| [internal/workers/user\_notifier.go](/internal/workers/user_notifier.go) | Go | 65 | 24 | 8 | 97 |
| [llmsvcsettings.yaml](/llmsvcsettings.yaml) | YAML | 37 | 0 | 12 | 49 |
| [migrations/0\_memotogo\_tables.down.sql](/migrations/0_memotogo_tables.down.sql) | MS SQL | 3 | 0 | 0 | 3 |
| [migrations/20260313130202\_memotogo\_tables.down.sql](/migrations/20260313130202_memotogo_tables.down.sql) | MS SQL | 3 | 0 | 1 | 4 |
| [migrations/20260313130202\_memotogo\_tables.up.sql](/migrations/20260313130202_memotogo_tables.up.sql) | MS SQL | 39 | 1 | 8 | 48 |
| [migrations/20260325140617\_memotogo\_tables.down.sql](/migrations/20260325140617_memotogo_tables.down.sql) | MS SQL | 1 | 0 | 0 | 1 |
| [migrations/20260325140617\_memotogo\_tables.up.sql](/migrations/20260325140617_memotogo_tables.up.sql) | MS SQL | 15 | 0 | 1 | 16 |
| [migrations/migrate.sh](/migrations/migrate.sh) | Shell Script | 1 | 0 | 0 | 1 |

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)