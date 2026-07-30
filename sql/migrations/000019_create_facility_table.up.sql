CREATE TABLE facility (
    id VARCHAR(4) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(255) NOT NULL DEFAULT ''
);

INSERT INTO facility (id, name, url) VALUES
    ('ZAB', 'Albuquerque ARTCC', 'https://www.zabartcc.org/'),
    ('ZAN', 'Anchorage ARTCC', 'https://www.zanartcc.org/'),
    ('ZTL', 'Atlanta ARTCC', 'https://www.ztlartcc.org/'),
    ('ZBW', 'Boston ARTCC', 'http://www.bvartcc.com/'),
    ('ZAU', 'Chicago ARTCC', 'https://www.zauartcc.org/'),
    ('ZOB', 'Cleveland ARTCC', 'https://clevelandcenter.org/'),
    ('ZDV', 'Denver ARTCC', 'https://zdvartcc.org/'),
    ('ZFW', 'Fort Worth ARTCC', 'https://www.zfwartcc.net/'),
    ('HCF', 'Honolulu', 'https://vhcf.net/'),
    ('ZHU', 'Houston ARTCC', 'https://houston.center/'),
    ('ZID', 'Indianapolis ARTCC', 'https://flyindycenter.com/'),
    ('ZJX', 'Jacksonville ARTCC', 'https://zjxartcc.org/'),
    ('ZKC', 'Kansas City ARTCC', 'https://vzkc.org/'),
    ('ZLA', 'Los Angeles ARTCC', 'https://laartcc.org/'),
    ('ZME', 'Memphis ARTCC', 'https://memphisartcc.com/'),
    ('ZMA', 'Miami ARTCC', 'https://www.zmaartcc.net/'),
    ('ZMP', 'Minneapolis ARTCC', 'https://minniecenter.org/'),
    ('ZNY', 'New York ARTCC', 'https://nyartcc.org/'),
    ('ZOA', 'Oakland ARTCC', 'https://oakartcc.org/'),
    ('ZLC', 'Salt Lake City ARTCC', 'https://zlcartcc.org/'),
    ('ZSE', 'Seattle ARTCC', 'https://zseartcc.org/'),
    ('ZDC', 'Washington, D.C. ARTCC', 'https://www.vzdc.org/');
