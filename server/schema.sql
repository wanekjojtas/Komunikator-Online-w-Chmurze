-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Table: public.chat_members
CREATE TABLE public.chat_members (
    chat_id uuid NOT NULL,
    user_id uuid NOT NULL,
    PRIMARY KEY (chat_id, user_id),
    FOREIGN KEY (chat_id) REFERENCES public.chats(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES public.users(id)
);

-- Table: public.chats
CREATE TABLE public.chats (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255),
    creator_id uuid,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (creator_id) REFERENCES public.users(id)
);

-- Table: public.messages
CREATE TABLE public.messages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chat_id uuid,
    sender_id uuid,
    content text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (chat_id) REFERENCES public.chats(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES public.users(id)
);

-- Table: public.schema_migrations
CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL,
    PRIMARY KEY (version)
);

-- Table: public.users
CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    username character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    password text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE (email),
    UNIQUE (username)
);
