CREATE TABLE IF NOT EXISTS `category` (
    id int(10) unsigned NOT NULL AUTO_INCREMENT,
    name varchar(255) COLLATE utf8_unicode_ci NOT NULL,
    slug varchar(255) NOT NULL,
    created_at timestamp NULL DEFAULT NULL  ,
    updated_at timestamp NULL DEFAULT NULL ,
    PRIMARY KEY (id),
    UNIQUE (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
 