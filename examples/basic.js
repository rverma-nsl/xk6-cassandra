import {randomIntBetween, randomString} from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import cql from 'k6/x/cassandra';

// CREATE KEYSPACE ge WITH replication = {'class': 'NetworkTopologynumategy', 'AWS_AP_SOUTH_1' : 3} AND durable_writes = true;
// CREATE KEYSPACE ge WITH replication = {'class': 'SimpleStrategy', 'replication_factor' : 1} AND durable_writes = true;
const keyspace = 'ge'
const db = cql.connect({
    url: "localhost:9042",
    username: "scylla",
    password: "xxxx",
    dc: "AWS_AP_SOUTH_1"
});

export function setup() {
    let exists = cql.checkTable(db, keyspace, 'record')
    if (!exists) {
        // console.log(`create shard table`)
        // cql.exec(db, `CREATE TABLE IF NOT EXISTS ge.shard_table (
        //     id uuid,
        //     shard_id uuid,
        //     PRIMARY KEY(id, shard_id));`)
        console.log(`table doesn't exist in the keyspace ${keyspace}`)
        cql.exec(db, `CREATE TABLE IF NOT EXISTS ge.record
                      (
                          tenant          TEXT,
                          ge              TEXT,
                          idstr1          TEXT,
                          shard_id        INT,
                          insert_time TIMESTAMP,
                          idstr2          TEXT,
                          idstr3          TEXT,
                          idstr4          TEXT,
                          str1            TEXT,
                          str2            TEXT,
                          str3            TEXT,
                          str4            TEXT,
                          str5            TEXT,
                          str6            TEXT,
                          str7            TEXT,
                          str8            TEXT,
                          str9            TEXT,
                          str10           TEXT,
                          PRIMARY KEY ((tenant, ge,shard_id),idstr1) 
                      )WITH CLUSTERING ORDER BY (idstr1 ASC);`
        )
        cql.exec(db, `create index on ge.record ((tenant,ge,shard_id),idstr2);`)
    }
}

export default function () {
    let col = ["tenant", "ge","shard_id","idstr1","insert_time"]
    const ge = randomIntBetween(8, 99)
    const tenant = randomIntBetween(10000, 10999)
    const geBit = dec2bin(ge).slice(-3).split('').map(Number)
    const tenantBit = dec2bin(tenant).slice(-10).split('').map(Number)
    let vals = [`'ten_${tenant}'`, `'ge_${ge}'`, 0 , `'${randomString(20)}'`,Date.now()]
    for (let i = 2; i <= 4; i++) {
        if (geBit[i - 2] === 1) {
            col.push(`idstr${i}`)
            vals.push(`'${randomString(20)}'`)
        }
    }
    for (let i = 1; i <= 10; i++) {
        if (tenantBit[i - 1] === 1) {
            col.push(`str${i}`)
            vals.push(`'${randomString(20)}'`)
        }
    }
    cql.insert(db, keyspace, 'record', col, vals)
}

export function teardown() {
    db.close();
}

function dec2bin(dec) {
    return (dec >>> 0).toString(2);
}