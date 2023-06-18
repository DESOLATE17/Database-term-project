CREATE EXTENSION IF NOT EXISTS CITEXT;

CREATE UNLOGGED TABLE users
(
    Nickname CITEXT PRIMARY KEY,
    FullName TEXT NOT NULL,
    About    TEXT NOT NULL DEFAULT '',
    Email    CITEXT UNIQUE
);

CREATE UNLOGGED TABLE forum
(
    Title   TEXT NOT NULL,
    "user"  CITEXT,
    Slug    CITEXT PRIMARY KEY,
    Posts   INT DEFAULT 0,
    Threads INT DEFAULT 0
);

CREATE UNLOGGED TABLE thread
(
    Id      SERIAL PRIMARY KEY,
    Title   TEXT NOT NULL,
    Author  CITEXT REFERENCES "users" (Nickname),
    Forum   CITEXT REFERENCES "forum" (Slug),
    Message TEXT NOT NULL,
    Votes   INT                      DEFAULT 0,
    Slug    CITEXT,
    Created TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE UNLOGGED TABLE post
(
    Id       SERIAL PRIMARY KEY,
    Author   CITEXT,
    Created  TIMESTAMP WITH TIME ZONE DEFAULT now(),
    Forum    CITEXT,
    IsEdited BOOLEAN                  DEFAULT FALSE,
    Message  CITEXT NOT NULL,
    Parent   INT                      DEFAULT 0,
    Thread   INT,
    Path     INTEGER[],
    FOREIGN KEY (thread) REFERENCES "thread" (id),
    FOREIGN KEY (author) REFERENCES "users" (nickname)
);

CREATE UNLOGGED TABLE vote
(
    ID     SERIAL PRIMARY KEY,
    Author CITEXT REFERENCES "users" (Nickname),
    Voice  INT NOT NULL,
    Thread INT,
    FOREIGN KEY (thread) REFERENCES "thread" (id),
    UNIQUE (Author, Thread)
);


CREATE UNLOGGED TABLE users_forum
(
    Nickname CITEXT NOT NULL,
    FullName TEXT   NOT NULL,
    About    TEXT,
    Email    CITEXT,
    Slug     CITEXT NOT NULL,
    FOREIGN KEY (Nickname) REFERENCES "users" (Nickname),
    FOREIGN KEY (Slug) REFERENCES "forum" (Slug),
    UNIQUE (Nickname, Slug)
);

CREATE OR REPLACE FUNCTION updatePostUsersForum() RETURNS TRIGGER AS
$update_forum_posts$
DECLARE
    t_fullname CITEXT;
    t_about    CITEXT;
    t_email    CITEXT;
BEGIN
    SELECT fullname, about, email FROM users WHERE nickname = NEW.author INTO t_fullname, t_about, t_email;
    INSERT INTO users_forum (nickname, fullname, about, email, Slug)
    VALUES (New.Author, t_fullname, t_about, t_email, NEW.forum)
    on conflict do nothing;
    return NEW;
end
$update_forum_posts$ LANGUAGE plpgsql;

CREATE TRIGGER post_user_forum
    AFTER INSERT
    ON post
    FOR EACH ROW
EXECUTE PROCEDURE updatePostUsersForum();

CREATE OR REPLACE FUNCTION updateThreadUserForum() RETURNS TRIGGER AS
$update_user_forum$
DECLARE
    a_nick     CITEXT;
    t_fullname CITEXT;
    t_about    CITEXT;
    t_email    CITEXT;
BEGIN
    SELECT Nickname, fullname, about, email
    FROM users
    WHERE Nickname = new.Author
    INTO a_nick, t_fullname, t_about, t_email;
    INSERT INTO users_forum (nickname, fullname, about, email, slug)
    VALUES (a_nick, t_fullname, t_about, t_email, NEW.forum)
    on conflict do nothing;
    return NEW;
end
$update_user_forum$ LANGUAGE plpgsql;

CREATE TRIGGER update_user_forum
    AFTER INSERT
    ON thread
    FOR EACH ROW
EXECUTE PROCEDURE updateThreadUserForum();


CREATE OR REPLACE FUNCTION updateCountOfThreads() RETURNS TRIGGER AS
$update_forum_threads$
BEGIN
    UPDATE forum SET Threads=(Threads + 1) WHERE slug = NEW.forum;
    return NEW;
end
$update_forum_threads$ LANGUAGE plpgsql;

CREATE TRIGGER update_forum
    BEFORE INSERT
    ON thread
    FOR EACH ROW
EXECUTE PROCEDURE updateCountOfThreads();

CREATE OR REPLACE FUNCTION updateVotes() RETURNS TRIGGER AS
$update_votes$
BEGIN
    IF (TG_OP = 'UPDATE') THEN
        IF OLD.Voice <> NEW.Voice THEN
            UPDATE thread SET votes=(votes + NEW.Voice * 2) WHERE id = NEW.Thread;
        END IF;
        return NEW;
    ELSIF (TG_OP = 'INSERT') THEN
        UPDATE thread SET votes=(votes + NEW.voice) WHERE id = NEW.thread;
        return NEW;
    END IF;
end
$update_votes$ LANGUAGE plpgsql;

CREATE TRIGGER update_votes
    BEFORE UPDATE OR INSERT
    ON vote
    FOR EACH ROW
EXECUTE PROCEDURE updateVotes();

CREATE OR REPLACE FUNCTION updatePath() RETURNS TRIGGER AS
$update_path$
DECLARE
    parent_path   INTEGER[];
    parent_thread int;
BEGIN
    IF (NEW.parent = 0) THEN
        NEW.path := array_append(new.path, new.id);
    ELSE
        SELECT thread FROM post WHERE id = new.parent INTO parent_thread;
        IF NOT FOUND OR parent_thread != NEW.thread THEN
            RAISE EXCEPTION 'NOT FOUND OR parent_thread != NEW.thread' USING ERRCODE = '22409';
        end if;

        SELECT path FROM post WHERE id = new.parent INTO parent_path;
        NEW.path := parent_path || new.id;
    END IF;
    UPDATE forum SET Posts=Posts + 1 WHERE forum.slug = new.forum;
    RETURN new;
END
$update_path$ LANGUAGE plpgsql;

CREATE TRIGGER update_path
    BEFORE INSERT
    ON post
    FOR EACH ROW
EXECUTE PROCEDURE updatePath();
