--SET search_path TO memotogo;
 
CREATE TABLE users (
  id serial PRIMARY KEY,
  username text UNIQUE NOT NULL,
  chat_id integer, 
  approved boolean NOT NULL,
  welcome_msg_sent timestamp, 
  created_at timestamp DEFAULT now()
);

CREATE TABLE meetings (
  id bigserial PRIMARY KEY,
  date timestamp DEFAULT now(),
  duration integer DEFAULT 0,
  transcription text,
  summary text,
  topics jsonb,
  search_vector_en tsvector,
  search_vector_ru tsvector,
  user_id integer NOT NULL REFERENCES users(id)
);

CREATE FUNCTION meetings_search_update() RETURNS trigger AS $$
BEGIN
  IF NEW.transcription IS DISTINCT FROM OLD.transcription
    OR NEW.summary IS DISTINCT FROM OLD.summary THEN

    NEW.search_vector_en :=
      to_tsvector('english', coalesce(NEW.transcription, '') || ' ' || coalesce(NEW.summary, ''));
    NEW.search_vector_ru :=
      to_tsvector('russian', coalesce(NEW.transcription, '') || ' ' || coalesce(NEW.summary, ''));
  END IF;

  RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER tsvector_update
BEFORE INSERT OR UPDATE ON meetings
FOR EACH ROW EXECUTE FUNCTION meetings_search_update();


CREATE INDEX idx_meetings_search_vector_en ON meetings USING GIN(search_vector_en);
CREATE INDEX idx_meetings_search_vector_ru ON meetings USING GIN(search_vector_ru);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_meetings_user_id ON meetings(user_id);
CREATE INDEX idx_meetings_summary ON meetings(summary);