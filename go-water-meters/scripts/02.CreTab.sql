CREATE TABLE wmeter.page (
    page_id       serial PRIMARY KEY,
	page_title    varchar(64),
	page_template varchar(64),
	controller    varchar(64),
	action        varchar(64),
	page_url      varchar(256),
	constraint page_url_uk unique (page_url)
);

CREATE TABLE wmeter.request (
    request_id        serial PRIMARY KEY,
	request_url       varchar(256),
	controller        varchar(64),
	action            varchar(64),
	redirect_url      varchar(256),
	redirect_on_error varchar(256)
	constraint request_url_uk unique (request_url)
);
