package model

import (
    "github.com/jinzhu/gorm"
    _"github.com/jinzhu/gorm/dialects/sqlite"
    "golang.org/x/crypto/bcrypt"
    jwt "github.com/dgrijalva/jwt-go"
    // "github.com/satori/go.uuid"
    "fmt"
    //"bytes"
    // "time"
    //"errors"
)

const TypedHello string = "Hello, 世界"
var db *gorm.DB

type News struct {
    gorm.Model
    Title    string `gorm:"size:255"`
    URL      string
    Hash    string `gorm:"size:255"`
    Audios []Audio
}

type Audio struct {
    gorm.Model
    NewsId string
    UserId string //`gorm:"ForeignKey:RestaurantID";AssociationForeignKey:ReferenceRoll`
    URL string
    Length int64
    Price float64
}

type User struct {
    gorm.Model
    Name string
    Password string
    Wallet Wallet
}

type Wallet struct {
    gorm.Model
    Ballance float64
}

func init() {
    var initErr error
    db, initErr = gorm.Open("sqlite3", "public/db.sqlite")
    //db, initErr = gorm.Open("postgres", "host=localhost user=admin dbname=squadread sslmode=disable password=pass")
    if initErr != nil {
        panic("failed to connect database")
    }

    if (db.HasTable(&Wallet{})) {
        db.DropTable(&Wallet{})
    }

    if (db.HasTable(&Audio{})) {
        db.DropTable(&Audio{})
    }

    if (db.HasTable(&User{})) {
        db.DropTable(&User{})
    }

    if (db.HasTable(&News{})) {
        db.DropTable(&News{})
    }

    db.CreateTable(&Wallet{})
    db.CreateTable(&Audio{})
    db.CreateTable(&User{})
    db.CreateTable(&News{})
}

func AddUser(userName string, userPassword string) (string, error) {
    password := []byte(userPassword)

    // Hashing the password with the default cost of 10
    hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
    if err != nil {
        panic(err)
    }

    fmt.Println(string(hashedPassword))

    // Comparing the password with the hash
    err = bcrypt.CompareHashAndPassword(hashedPassword, password)
    if err == nil {
       // n := bytes.Index(hashedPassword, []byte{0})

        u := User {
            Name: userName,
            Password:  string(hashedPassword[:]),
        }
        db.Debug().Set("gorm:save_associations", true).Create(&u)
        
        tokenString, err := u.createToken()
        if err == nil {
            return tokenString, nil
        }
    }
  return "", err
}

func AuthoriseUser(userName string, userPassword string) (string, error) {
    password := []byte(userPassword)

    // Hashing the password with the default cost of 10
    hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
    if err != nil {
        panic(err)
    }

    fmt.Println(string(hashedPassword))
    // Comparing the password with the hash
    err = bcrypt.CompareHashAndPassword(hashedPassword, password)
    if err == nil {
       // n := bytes.Index(hashedPassword, []byte{0})
        u := User{}
        db.Where("name = ? ",userName ).First(&u)
        fmt.Println("here 0")
        if u.Name != "" {
           tokenString, err := u.createToken()
            if err == nil {
                return tokenString, nil
            }
        }
        
    }
    return "", err
}


func (u *User) createToken() (tokenString string,err error) {
    claims := &jwt.StandardClaims{
        ExpiresAt: 15000,
        Issuer: u.Name,
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err = token.SignedString([]byte("darksecret"))

    return tokenString, err
}


func ValidateToken(tokenString string) User {

return User{}
    // token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    //     // Don't forget to validate the alg is what you expect:
    //     if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
    //         return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
    //     }
    //     return hmacSampleSecret, nil
    // })

    // if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
    //     fmt.Println(claims["Issuer"])
    // } else {
    //     fmt.Println(err)
    // }

}