package pagenation

import (
	"fmt"
	"reflect"
	"strings"
	"github.com/go-xorm/xorm"
	"strconv"
)

type DaoBase struct {
	Datasource *xorm.Engine
}

/**
	这里使用xorm引擎
	"github.com/go-xorm/xorm"
	_ "github.com/go-sql-driver/mysql"
 */
func (DaoBase *DaoBase) GetDatasource() *xorm.Engine {
	return DaoBase.Datasource
}

/**
	这里使用xorm引擎
	"github.com/go-xorm/xorm"
	_ "github.com/go-sql-driver/mysql"
 */
func (DaoBase *DaoBase) SetDatasource(datasource *xorm.Engine) {
	DaoBase.Datasource = datasource
}

/**
	分页
	注意 ： 连接需要传入datasource
 */
func (DaoBase *DaoBase) GetPageLists(po interface{}, table string,fields string ,pk string, alias string,join string,condition string, order string,group string, page int, listRow int) map[string]interface{} {
	var _fields = "*"
	var _page = 0
	var _order = ""
	var orderCmd = " ORDER BY "
	var groupBy  = " GROUP BY "
	var _group   = ""
	var _pk = "`id`"
	var count = 0
	var _listRow = LIST_ROWS
	if pk == "" {
		_pk = pk
	}
	if fields != ""{
		_fields = fields
	}
	var condi = " "
	if condition != "" {
		condi = condition
	}
	if order != "" {
		_order = orderCmd + order + " "
	} else {
		_order += orderCmd + _pk + " " + " DESC "
	}
	if group != ""{
		_group = groupBy + group
	}
	if page >= 0 {
		_page = page
	}
	if listRow > 0{
		_listRow = listRow
	}
	var querySql = fmt.Sprintf("SELECT %s FROM %s "+ " "+alias+" "+ join + " WHERE 1 ", _fields, table)
	var countSql = fmt.Sprintf("SELECT count(%s) AS count FROM %s "+ " "+alias+" "+ join + " WHERE 1 ", pk, table)
	querySql = querySql + condi + _group + _order
	countSql = countSql + condi + _group
	var handler = DaoBase.GetDatasource()
	countRes, _ := handler.QueryString(countSql)
	handler.SQL(QueryBuild(querySql, _page, _listRow,true)).Find(po)
	var resData = make(map[string]interface{}, 0)
	if len(countRes) > 0{
		count, _ = strconv.Atoi(countRes[0]["count"])
	}
	if count > 0 {
		resData["list"] = po
	} else {
		resData["list"] = make([]int, 0)
	}
	var pageInfo = CommaPaginator(_page, listRow, int64(count))
	resData["total_page"] = pageInfo["total_page"]
	resData["curr_page"] = pageInfo["curr_page"]
	resData["page_rows"] = pageInfo["page_rows"]
	resData["total_record"] = count
	return resData
}

/**
	不分页
	注意 ： 连接需要传入datasource
 */
func (DaoBase *DaoBase) GetLists(po interface{}, table string,fields string ,pk string,alias string,join string,condition string, order string,group string) map[string]interface{} {
	var _fields = "*"
	var _order = ""
	var orderCmd = " ORDER BY "
	var groupBy  = " GROUP BY "
	var _group   = ""
	var _pk = "`id`"
	if pk == "" {
		_pk = pk
	}

	var condi = " "
	if condition != "" {
		condi = condition
	}
	if order != "" {
		_order = orderCmd + order + " "
	} else {
		_order += orderCmd + _pk + " " + " DESC "
	}
	if group != ""{
		_group = groupBy + group
	}
	var querySql = fmt.Sprintf("SELECT %s FROM %s "+ " "+alias+" "+ join + " WHERE 1 ", _fields, table)
	querySql = querySql + condi + _group + _order
	var handler = DaoBase.GetDatasource()
	handler.SQL(querySql).Find(po)
	var resData = make(map[string]interface{}, 0)
	resData["list"] = po
	return resData
}



/**
	测试用例
	var data = [][]string{
		{"INNER","b b","b.id = a.id"},
		{"INNER","c c","c.id = b.id"},
	}
	var base = &DaoBase{}
	fmt.Print(base.ConditionJoin(data))

	Join 条件生成器
	authro : Bill
 */
func (DaoBase *DaoBase) ConditionJoin(join [][]string) string{
	var _join = ""
	var (
		_innerFlag = "INNER"
		_leftFlag =  "LEFT"
		_rightFlag = "RIGHT"
	)
	var innerTpl = " INNER JOIN %s ON %s "
	var leftTpl  = " LEFT JOIN %s ON %s "
	var rightTpl  = " RIGHT JOIN %s ON %s "
	for _, row := range join {
		var in = strings.ToUpper(strings.TrimSpace(row[0]))
		var table = row[1]
		var condi = row[2]
		switch in {
			case _innerFlag:
				_join += fmt.Sprintf(innerTpl,table,condi)
			case _leftFlag:
				_join += fmt.Sprintf(leftTpl,table,condi)
			case _rightFlag:
				_join += fmt.Sprintf(rightTpl,table,condi)
			default:
				_join += fmt.Sprintf(innerTpl,table,condi)
		}
	}
	return _join
}


/**
	SqlBuildConditon : 测试用例
		var condi = make(map[string]map[string]interface{})
		var inCodi = make(map[string]interface{})
		var like = make(map[string]interface{})
		var null = make(map[string]interface{})        // null
		var or = make(map[string]interface{},0)

		orCondi["condi"]  = fmt.Sprintf("bc.status = %d",0)
		nullCondi["bc.cv_id"] = nil
		like["name"] = "Bill"
		inCodi["type"] = 1
		inCodi["title"] = "yang"
		condi["AND"] = inCodi
		condi["LIKE"] = like
		condi["OR"] = or
		condi["NULL"] = null
		fmt.Print(ConditionBuild(condi))
	author ： Bill
 */
func (DaoBase *DaoBase) ConditionBuild(condi map[string]map[string]interface{}) string {
	if len(condi) == 0 {
		return ""
	}
	//flag部分
	var (
		_andFlag  = "AND"
		_likeFlag = "LIKE"
		_gtFlag   = "GT" //比较
		_ltFlag   = "LT" //比较
		_InFlag   = "IN" //比较
		_nullFlag = "NULL" //为空
		_orFlag   = "OR"   //或
	)

	var _condi = ""
	var str = " AND ( %s = '%v' )"
	var intCondi = " AND ( %s = %v )"
	var _gtCondi = " AND ( %s > %v )"
	var _ltCondi = " AND ( %s < %v )"
	var _InCondi = " AND ( %s IN (%s) )"
	var _NullCondi = " AND ( ISNULL(%s) )"
	var _orCondi = " OR ( %v )"
	var _currRel = ""
	for _rela, v := range condi {
		_currRel = strings.ToUpper(_rela)
		for _field, _v := range v {
			switch _currRel {
			case _andFlag:
				if reflect.TypeOf(_v).String() == "string" {
					_condi += fmt.Sprintf(str, _field, _v)
				} else {
					_condi += fmt.Sprintf(intCondi, _field, _v)
				}
			case _likeFlag:
				_condi += " AND ( " + _field + " LIKE '%" + _v.(string) + "%' )"
			case _gtFlag:
				_condi += fmt.Sprintf(_gtCondi, _field, _v)
			case _ltFlag:
				_condi += fmt.Sprintf(_ltCondi, _field, _v)
			case _InFlag:
				_condi += fmt.Sprintf(_InCondi, _field, _v)
			case _nullFlag:
				_condi += fmt.Sprintf(_NullCondi, _field)
			case _orFlag:
				_condi += fmt.Sprintf(_orCondi, _v)
			}

		}
	}
	return _condi
}

func (DaoBase *DaoBase) GetByPo(po interface{},table string,condi string) interface{}{
	var handler = DaoBase.GetDatasource()
	handler.SQL(fmt.Sprintf("SELECT * FROM %s WHERE 1 " + condi,table)).Find(po)
	return po
}

func (DaoBase *DaoBase) EditRow(table string,condi string,params map[string]interface{})int{
	var handler = DaoBase.GetDatasource()
	effRow,_ := handler.Table(table).Where(condi).Update(params)
	return int(effRow)
}

func (DaoBase *DaoBase)InsertRow(table string,params map[string]interface{})(int,bool){
	if len(params) == 0{
		return 0,false
	}
	var (
		sep = ","
		valBuff = ""
		_fields = make([]string,0)
		_vals = make([]interface{},0)
	)
	for key,row := range params{
		_fields = append(_fields,key )
		_vals = append(_vals,row )
	}
	for _,row := range _vals{
		var val = ""
		if strings.Contains(reflect.TypeOf(row).String(),"string"){
			val = fmt.Sprintf("'%v'",row)
		}else{
			val = fmt.Sprintf("%v",row)
		}
		valBuff += val + sep
	}
	var handler = DaoBase.GetDatasource()
	sql := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)",table,strings.Join(_fields,sep),strings.Trim(valBuff,sep))
	res,err := handler.Exec(sql)
	if err != nil{
		return 0 ,false
	}
	nId,err := res.LastInsertId()
	if err != nil{
		return 0 ,false
	}
	return int(nId),true
}
