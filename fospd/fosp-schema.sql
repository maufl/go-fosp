--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: data; Type: TABLE; Schema: public; Owner: fosp; Tablespace: 
--

CREATE TABLE data (
    id bigint NOT NULL,
    uri text,
    parent_id bigint NOT NULL,
    content text
);


ALTER TABLE public.data OWNER TO fosp;

--
-- Name: data_id_seq; Type: SEQUENCE; Schema: public; Owner: fosp
--

CREATE SEQUENCE data_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.data_id_seq OWNER TO fosp;

--
-- Name: data_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: fosp
--

ALTER SEQUENCE data_id_seq OWNED BY data.id;


--
-- Name: data_parent_id_seq; Type: SEQUENCE; Schema: public; Owner: fosp
--

CREATE SEQUENCE data_parent_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.data_parent_id_seq OWNER TO fosp;

--
-- Name: data_parent_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: fosp
--

ALTER SEQUENCE data_parent_id_seq OWNED BY data.parent_id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: fosp; Tablespace: 
--

CREATE TABLE users (
    name character varying(256),
    password character varying(256)
);


ALTER TABLE public.users OWNER TO fosp;

--
-- Name: id; Type: DEFAULT; Schema: public; Owner: fosp
--

ALTER TABLE ONLY data ALTER COLUMN id SET DEFAULT nextval('data_id_seq'::regclass);


--
-- Name: parent_id; Type: DEFAULT; Schema: public; Owner: fosp
--

ALTER TABLE ONLY data ALTER COLUMN parent_id SET DEFAULT nextval('data_parent_id_seq'::regclass);


--
-- Name: data_pkey; Type: CONSTRAINT; Schema: public; Owner: fosp; Tablespace: 
--

ALTER TABLE ONLY data
    ADD CONSTRAINT data_pkey PRIMARY KEY (id);


--
-- Name: data_uri_key; Type: CONSTRAINT; Schema: public; Owner: fosp; Tablespace: 
--

ALTER TABLE ONLY data
    ADD CONSTRAINT data_uri_key UNIQUE (uri);


--
-- Name: users_name_key; Type: CONSTRAINT; Schema: public; Owner: fosp; Tablespace: 
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_name_key UNIQUE (name);


--
-- Name: public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

