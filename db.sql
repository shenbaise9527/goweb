create table user(
    id         int(11) not null primary key auto_increment, 
    uuid       varchar(64) not null, 
    name       varchar(255), 
    email      varchar(255) not null, 
    password   varchar(255) not null, 
    created_at datetime not null,
    unique key uk_email(email)
)ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET='utf8';

create table sessions (
    id         int(11) not null primary key auto_increment, 
    uuid       varchar(64) not null, 
    email      varchar(255), 
    user_id    int(11), 
    created_at datetime not null
)ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET='utf8';

create table threads (
    id         int(11) not null primary key auto_increment, 
    uuid       varchar(64) not null,
    topic      text,
    user_id    int(11), 
    created_at datetime not null
)ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET='utf8';

create table posts (
    id         int(11) not null primary key auto_increment, 
    uuid       varchar(64) not null, 
    body       text,
    user_id    int(11), 
    thread_id  int(11), 
    created_at datetime not null
)ENGINE=INNODB AUTO_INCREMENT=1 DEFAULT CHARSET='utf8';

