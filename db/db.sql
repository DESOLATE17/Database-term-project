CREATE EXTENSION IF NOT EXISTS CITEXT; -- eliminate calls to lower

CREATE UNLOGGED TABLE users
(
    Nickname   CITEXT PRIMARY KEY,
    FullName   TEXT NOT NULL,
    About      TEXT NOT NULL DEFAULT '',
    Email      CITEXT UNIQUE
);

CREATE UNLOGGED TABLE forum
(
    Title    TEXT   NOT NULL,
    "user"   CITEXT,
    Slug     CITEXT PRIMARY KEY,
    Posts    INT    DEFAULT 0,
    Threads  INT    DEFAULT 0
);

CREATE UNLOGGED TABLE threads
(
    Id      SERIAL    PRIMARY KEY,
    Title   TEXT      NOT NULL,
    Author  CITEXT    REFERENCES "users"(Nickname),
    Forum   CITEXT    REFERENCES "forum"(Slug),
    Message TEXT      NOT NULL,
    Votes   INT       DEFAULT 0,
    Slug    CITEXT,
    Created TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE UNLOGGED TABLE posts
(
    Id        SERIAL      PRIMARY KEY,
    Author    CITEXT,
    Created   TIMESTAMP   WITH TIME ZONE DEFAULT now(),
    Forum     CITEXT,
    IsEdited  BOOLEAN     DEFAULT FALSE,
    Message   CITEXT      NOT NULL,
    Parent    INT         DEFAULT 0,
    Thread    INT,
    Path      INTEGER[],
    FOREIGN KEY (thread) REFERENCES "threads" (id),
    FOREIGN KEY (author) REFERENCES "users"  (nickname)
);

CREATE UNLOGGED TABLE votes
(
    ID       SERIAL PRIMARY KEY,
    Author   CITEXT    REFERENCES "users" (Nickname),
    Voice    INT       NOT NULL,
    Thread   INT,
    FOREIGN KEY (thread) REFERENCES "threads" (id),
    UNIQUE (Author, Thread)
);


CREATE UNLOGGED TABLE users_forum
(
    Nickname  CITEXT  NOT NULL,
    FullName  TEXT    NOT NULL,
    About     TEXT,
    Email     CITEXT,
    Slug      CITEXT  NOT NULL,
    FOREIGN KEY (Nickname) REFERENCES "users" (Nickname),
    FOREIGN KEY (Slug) REFERENCES "forum" (Slug),
    UNIQUE (Nickname, Slug)
);
