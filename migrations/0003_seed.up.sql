-- bcrypt hash of plaintext: password
INSERT INTO users
    (
        id           ,
        email        ,
        password_hash,
        display_name ,
        created_at
    )
VALUES
    (
        1                                                             ,
        'alice@example.com'                                           ,
        '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
        'Alice'                                                       ,
        datetime('now')
    ),
    (
        2                                                             ,
        'bob@example.com'                                             ,
        '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
        'Bob'                                                         ,
        datetime('now')
    )
    ,
    (
        3                                                             ,
        'carol@example.com'                                           ,
        '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
        'Carol'                                                       ,
        datetime('now')
    )
    ,
    (
        4                                                             ,
        'dave@example.com'                                            ,
        '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
        'Dave'                                                        ,
        datetime('now')
    )
    ,
    (
        5                                                             ,
        'eve@example.com'                                             ,
        '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
        'Eve'                                                         ,
        datetime('now')
    )
;
INSERT INTO posts
    (
        id     ,
        user_id,
        title  ,
        body   ,
        created_at
    )
VALUES
    (
        1                                   ,
        1                                   ,
        'POST title one'                    ,
        'THis is the body of the first post',
        datetime('now')
    )
    ,
    (
        2                 ,
        2                 ,
        'Post two title'  ,
        'Body for post 2.',
        datetime('now')
    )
    ,
    (
        3                 ,
        3                 ,
        'Post three title',
        'Body for post 3.',
        datetime('now')
    )
    ,
    (
        4                 ,
        4                 ,
        'Post four title' ,
        'Body for post 4.',
        datetime('now')
    )
    ,
    (
        5                 ,
        5                 ,
        'Post five title' ,
        'Body for post 5.',
        datetime('now')
    )
    ,
    (
        6                 ,
        1                 ,
        'Post six title'  ,
        'Body for post 6.',
        datetime('now')
    )
    ,
    (
        7                 ,
        2                 ,
        'Post seven title',
        'Body for post 7.',
        datetime('now')
    )
    ,
    (
        8                 ,
        3                 ,
        'Post eight title',
        'Body for post 8.',
        datetime('now')
    )
    ,
    (
        9                 ,
        4                 ,
        'Post nine title' ,
        'Body for post 9.',
        datetime('now')
    )
    ,
    (
        10                 ,
        5                  ,
        'Post ten title'   ,
        'Body for post 10.',
        datetime('now')
    )
;
INSERT INTO comments
    (
        id            ,
        post_id       ,
        name          ,
        email         ,
        body          ,
        author_user_id,
        created_at
    )
VALUES
    (
        1                   ,
        1                   ,
        'Alice'             ,
        'alice@example.com' ,
        'Comment on post 1.',
        1                   ,
        datetime('now')
    )
    ,
    (
        2                           ,
        1                           ,
        'Bob'                       ,
        'bob@example.com'           ,
        'Another comment on post 1.',
        2                           ,
        datetime('now')
    )
    ,
    (
        3                   ,
        2                   ,
        'Carol'             ,
        'carol@example.com' ,
        'Comment on post 2.',
        3                   ,
        datetime('now')
    )
    ,
    (
        4                   ,
        3                   ,
        'Dave'              ,
        'dave@example.com'  ,
        'Comment on post 3.',
        4                   ,
        datetime('now')
    )
    ,
    (
        5                   ,
        4                   ,
        'Eve'               ,
        'eve@example.com'   ,
        'Comment on post 4.',
        5                   ,
        datetime('now')
    )
    ,
    (
        6                   ,
        5                   ,
        'Alice'             ,
        'alice@example.com' ,
        'Comment on post 5.',
        1                   ,
        datetime('now')
    );