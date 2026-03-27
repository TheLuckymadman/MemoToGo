CREATE TABLE transcription_tasks (
  id bigserial PRIMARY KEY, 
  chat_id integer, 
  transcription_task_id text UNIQUE,
  transcription_status text NOT NULL, -- NEW, RUNNING, DONE, ERROR
  task_status text NOT NULL,  
  retry_cnt integer DEFAULT 0,
  upload_file_id text,
  download_file_id text,
  duration integer DEFAULT 0,
  summary_status text,
  summary_created timestamp,
  created_at timestamp DEFAULT now(),
  updated_at timestamp DEFAULT now(),
  meeting_id bigint REFERENCES meetings(id)
);

CREATE INDEX idx_transcription_tasks_transcription_status ON transcription_tasks(transcription_status);
CREATE INDEX idx_transcription_tasks_task_status ON transcription_tasks(task_status);