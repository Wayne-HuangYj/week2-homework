package main

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path"
	"encoding/json"
	"io/ioutil"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

/**
	1. 我们在数据库操作的时候，比如 dao 层中当遇到一个 sql.ErrNoRows 的时候，是否应该 Wrap 这个 error，
	抛给上层。为什么，应该怎么做请写出代码？
	答：应该Wrap错误，因为dao层查询的时候需要上层提供一些参数，而如果根据这些参数得不到数据行的，有
	可能是由于参数不规范引起的，为了在记录日志的时候能够提供有效信息，应该将参数内容连同错误内容Wrap起来，
	交给上层去记录。

	我简单地建立了一个数据表create table user (id int primary key, name varchar(24));用于通过id查询名字name。
	当传入数据库不存在的id的时候就会报错sql.ErrNoRows，err := db.QueryRow(query).Scan(&id, &name)
	然后就将这个错误连同错误的id信息一同Wrap给上层，上层既可以拿到原始错误，也可以获取额外的错误信息

	func QueryNameById(db *sql.DB, id int) (string, error) {
		query := fmt.Sprintf("select * from user where id = '%d';", id)
		var name string
		err := db.QueryRow(query).Scan(&id, &name)
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("no id=%d", id))
		}
		return name, nil
	}

	除此之外，我还做了一些读取数据库配置文件的工作，如果文件路径不对，或者配置文件不是JSON格式的，
	都会把错误Wrap返回给上层。
*/

type MySQLConfig struct {
	User string
	Password string
	Host string
	Port int
	Database string
}

type Config struct {
	Mysql MySQLConfig
}

// 真正读取文件的函数
func ReadFile(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "open failed")
	}
	return data, nil
}

// 在当前文件夹下读取DB的配置文件，用于连接数据库
func ReadDBConfig() ([]byte, error) {
	home, err := os.Getwd()
	if err != nil {  // 因为调用的是库函数，所以返回的肯定是原始错误，因此可以用Wrap错误封装上去
		return nil, errors.Wrap(err, "failed to get work directory")
	}
	config, err := ReadFile(path.Join(home, "config.json"))
	return config, errors.WithMessage(err, "failed to read config")  //  如果err是nil，则WithMessage也是返回nil
}

// 解析JSON成自定义结构体
func Loads(j []byte, v interface{}) error {
	err := json.Unmarshal(j, v)
	if err != nil {
		return errors.Wrap(err, "failed to load json")
	}
	return nil
} 

// dao 层中当遇到一个 sql.ErrNoRows 的时候，是否应该 Wrap 这个 error，抛给上层
func QueryNameById(db *sql.DB, id int) (string, error) {
	query := fmt.Sprintf("select * from user where id = '%d';", id)
	var name string
	err := db.QueryRow(query).Scan(&id, &name)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("no id=%d", id))
	}
	return name, nil
}

func main() {
	// 1. 读取配置文件
	rawConfig, err := ReadDBConfig()
	if err != nil {
		fmt.Printf("\n%+v\n", err)  // 打印错误信息和堆栈信息
		os.Exit(1)
	}
	// 2. 转换配置文件格式
	config := &Config{}
	err = Loads(rawConfig, config) // 将二进制的配置文件信息加载成可读的结构体格式
	if err != nil {
		fmt.Printf("\n%+v\n", err)  // 打印错误信息和堆栈信息
		os.Exit(1)
	}
	// 3. 连接数据库
	sqlConfig := config.Mysql
	connection := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", sqlConfig.User, sqlConfig.Password, sqlConfig.Host, sqlConfig.Port, sqlConfig.Database)
	db, err := sql.Open("mysql", connection)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// 查询数据
	id := 1
	name, err := QueryNameById(db, id)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			fmt.Printf("\n%+v\n", err)
		} else {
			fmt.Println(err)
		}
		os.Exit(1)
	} 
	fmt.Println(name)
}