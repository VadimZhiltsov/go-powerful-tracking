package main

import (
    "log"
    "net/http"
    "crypto/sha1"
    "encoding/hex"
    "encoding/json"
    "strings"
    "net"

    "fmt"
    "database/sql"
    "os"

    "github.com/codegangsta/martini"
    "github.com/martini-contrib/render"
    "github.com/garyburd/redigo/redis"
    "github.com/oschwald/maxminddb-golang"
    _ "github.com/go-sql-driver/mysql"
)

var mysql *sql.DB

func getHexFromString(ua string) string {
    h := sha1.New()
    h.Write([]byte(ua))
    s := hex.EncodeToString(h.Sum(nil))

    return s
}

type Targeting struct {
    Os string `json:"os"`
    Model    string `json:"model"`
    Manufacturer string `json:"manufacturer"`
    DeviceType string `json:"device_type"`
    OsVersion  float64 `json:"os_version"`
}

type onlyCountry struct {
    Country struct {
        ISOCode string `maxminddb:"iso_code"`
    } `maxminddb:"country"`
}


func main() {
    martini.Env = martini.Prod

    var err error
    mysql, err = sql.Open("mysql", "CONNECTION_STRING") // this does not really open a new connection
    if err != nil {
        log.Fatalf("Error on initializing database connection: %s", err.Error())
    }



    mysql.SetMaxIdleConns(100)
    mysql.SetMaxOpenConns(100)


    logString := "{ \"ad_id\": %s, \"app_id\": 999, \"site_id\": %s, \"price\": %s, \"ua\": %s, \"ip\": %s, \"query_params\": \"%s\", \"device\": \"%s\", \"os\": \"%s\", \"country\": \"%s\", \"os_version\":  \"%s\", \"creative_id\": \"%s\", \"manufacturer\": \"%s\", \"device_model\": \"%s\" }"


    redisPool := redis.NewPool(func() (redis.Conn, error) {
        c, err := redis.Dial("tcp", "CONNECTION_STRING")

        if err != nil {
            return nil, err
        }
        c.Do("SELECT", 5)
        return c, err
    }, 50)

    defer redisPool.Close()


    geoDB, err := maxminddb.Open("GeoIP2-Country-Test.mmdb")
    if err != nil {
        log.Fatal(err)
    }

    f, _ := os.OpenFile("./data.txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0666) 


    m := martini.Classic()

    m.Map(redisPool)

    m.Use(render.Renderer())

    m.Get("/",  func(r render.Render,req *http.Request, pool *redis.Pool, params martini.Params) {
        req.URL.RawQuery = strings.Replace(req.URL.RawQuery, ";", "%3B", -1)
         
        query := req.URL.Query()
        ip := net.ParseIP(query.Get("ip"))
        ad_id := query.Get("ad_id")
        site_id := query.Get("site_id")
        price := query.Get("price")
        ua := query.Get("ua")

        if ua == "" || ad_id == "" || site_id == "" || query.Get("ip") == "" {
             r.JSON(400, map[string]interface{}{
                "status":  "ERR",
                "message": "Query is not valid"})
            
            log.Print("Query is not valid")
            return
        }


        var isActive string
        
        qr := fmt.Sprintf("SELECT Ad.active * Site.active as valid FROM `Site` LEFT JOIN `Ad` ON  Ad.id = %s  WHERE Site.id = %s", ad_id, site_id)
        
        rows, err := mysql.Query(qr)
        if err != nil {
            fmt.Printf("Database Error!")
            log.Fatal(err)
            return
        }
        rows.Next()
        rows.Scan(&isActive)


        defer rows.Close()
        

        if isActive == "0" {
            r.JSON(400, map[string]interface{}{
                "status":  "ERR",
                "message": "Site or campaign is not active"})

            log.Print("Site or campaign is not active")
            return
        }


        uaHash := getHexFromString(ua)

        c := pool.Get()
        defer c.Close()

        uaJSON, err := redis.String(c.Do("GET", uaHash))

        var uaData Targeting
        err = json.Unmarshal([]byte(uaJSON), &uaData)

        var record onlyCountry // Or any appropriate struct
        err = geoDB.Lookup(ip, &record)
        if err != nil {
            log.Fatal(err)
        }

        ISOCode := record.Country.ISOCode

        response, err := c.Do("EVALSHA", "37a822553f54cc24b6866dcc2e26ae19998f5322", 0, uaData.Os, uaData.OsVersion, uaData.Manufacturer, uaData.Model, uaData.DeviceType, ISOCode, ad_id)
        vd := response.([]interface{})[3].([]interface{});
        var sum int64 = 0

        for i := 0; i < len(vd); i++ {
            sum += vd[i].(int64)
        }

        if sum != 6 {
            log.Print("Targeting validation has been failed")
            r.JSON(400, map[string]interface{}{
                "status":  "ERR",
                "message": "Targeting validation has been failed"})

        } else {
            r.JSON(200, map[string]interface{}{
                "status": "OK",
                "value":  "All right!"})


            s := fmt.Sprintf( logString, ad_id, site_id, price, query, ua, query.Get("ip"), uaData.DeviceType, uaData.Os, ISOCode, uaData.OsVersion, query.Get("creative_id"), uaData.Manufacturer, uaData.Model)

            _, err = f.WriteString(s)

            
            if err != nil {
                log.Fatal(err)
            }

            f.Close()
        }
    })

    m.RunOnAddr(":8080")
}