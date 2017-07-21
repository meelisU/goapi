package main

import (
	"net/http"
	"log"
	"github.com/gorilla/mux"
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"encoding/json"

)

var db *sql.DB
var err error

type Player struct {
	playerId string
	funds    int
}

type Tourment struct {
	tourmendId int;
	deposit    int;
}

type Balance struct {
	PlayerId string `json:"playerId"`
	Balance int `json:"balance"`
}


type Backer struct {
	Id    string
	money int
}
type TourmentBacker struct {
	Backerid string
	ShareofMoney int
	isbacker bool
}
type DBbackers struct {
	Backers []TourmentBacker `json:"Backers`
}

func main() {
	db, err = sql.Open("mysql", "root:root@tcp(localhost:3306)/goapi")
	errorCheck(err)
	r := mux.NewRouter()
	//r.HandleFunc("/test", defaultHandler)
	r.HandleFunc("/take", withdrawalHandler).Methods("GET")
	r.HandleFunc("/fund", fundHandler).Methods("GET")
	r.HandleFunc("/resultTournament", resultsHandler).Methods("GET")
	r.HandleFunc("/announceTournament", announcmentsHandler).Methods("POST")
	r.HandleFunc("/balance", balanceHandler).Methods("GET")
	r.HandleFunc("/joinTournament", tourmentHandler).Methods("GET")
	r.HandleFunc("/reset", resetDBHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", r))
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	var player Player;
	rows, err := db.Query("SELECT playerid,funds FROM player where id=2")
	if rows.Next() {
		errorCheck(err)
	}

	fmt.Println(player);
}

func withdrawalHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("playerId") == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest);
		return
	} else {
		var playerid = r.URL.Query().Get("playerId");
		playerparam := r.URL.Query().Get("points");
		funds, err := strconv.Atoi(playerparam)
		errorCheck(err)
		player := Player{playerid, funds}
		var dbplayer = findUserFromDB(player, false)
		if (hasEnoughMoney(player, dbplayer)) {
			withdrawMoney(dbplayer, dbplayer.funds-player.funds)
		} else {
			http.Error(w, "Not enough funds", http.StatusInternalServerError);
		}

	}


}

func fundHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("playerId") == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest);
		return
	} else {
		var playerid = r.URL.Query().Get("playerId");
		playerparam := r.URL.Query().Get("points");
		funds, err := strconv.Atoi(playerparam)
		errorCheck(err)
		player := Player{playerid, funds}
		var dbplayer =findUserFromDB(player, true)
		addFundingToPlayer(dbplayer,calculatefullfund(dbplayer.funds,funds))

	}
}
func calculatefullfund(original int, new int) int {
	return  original+new;
}


func announcmentsHandler(w http.ResponseWriter, r *http.Request) {

	if (r.FormValue("tournamentId") == "" || r.FormValue("deposit") == "" ) {
		http.Error(w, "Bad Request", http.StatusBadRequest);
		return
	} else {
		if (sameIdTourment(r.FormValue("tournamentId"))) {
			http.Error(w, "Already registered Id", http.StatusInternalServerError);
			return
		} else {
			var tourmentid string="T"
			tourmentIdvar := r.FormValue("tournamentId")
			depositvar := r.FormValue("deposit")
			tourmentid=tourmentid+ string(tourmentIdvar)
			deposit, err := strconv.Atoi(depositvar)
			errorCheck(err)

			createNewTourment(tourmentid, deposit);
		}
	}
}



func resultsHandler(w http.ResponseWriter, r *http.Request) {
	var winner Winner=findSumForTourment()
	var winningResults=startdivingingMoney(winner)
	js, err := json.Marshal(winningResults)
	errorCheck(err)
	w.Write(js)
}

type WinningResults struct {
	TourmentId string `json:"tournamentId":"`
	WinningPlayers [] WinningPlayer`json:"winners:"`
}
type WinningPlayer struct {
	PlayeriD string `json:"playerId"`
	Money string `json:"balance"`
}

func startdivingingMoney(winner Winner) WinningResults {
	w:=WinningResults{}
	if(winner.backers==""){

	}else{
		res :=DBbackers{}
		json.Unmarshal([]byte(winner.backers), &res)
		winners:=fundbackersAndPlayer(winner.playerid,res.Backers,winner.total)
		w=WinningResults{winner.tournamentId,winners}
	}
	return w;

}
func fundbackersAndPlayer(playerid string,backers []TourmentBacker,total int)[]WinningPlayer {

	var winners [] WinningPlayer
	for _, elem := range backers {

		f:=float64(elem.ShareofMoney)*0.25

		m:=float64(elem.ShareofMoney)+f
		total=total-int(m)
		tf:=getFundsForPlayer(elem.Backerid)+int(m);
		pl:=findUserById((string(elem.Backerid)))
		addFundingToPlayer(pl,tf)
		//var s string=strconv.Itoa((getFundsForPlayer(pl.playerId)))
		//w:=WinningPlayer{string(elem.Backerid),s }
		//winners=append(winners, w)

	}
	p1:=findUserById(playerid)
	addFundingToPlayer(p1,getFundsForPlayer(p1.playerId)+total)
	w:=WinningPlayer{p1.playerId,strconv.Itoa(getFundsForPlayer(p1.playerId))}
	winners=append(winners, w)
	return winners;
}




type Person struct {
	Id int
	Name string
}

type Winner struct {
	tournamentId  string
	playerid string
	total int
	backers string
}

func findSumForTourment()  Winner{
	//w :=Winner{}
	w := []Winner{}
	rows,err := db.Query("select  t.tournamentId, tp.playerid  ,SUM(tp.deposit) as total ,tp.backers from tourment as t left join (select tournamentId,playerid,deposit,backers from  tourmentdeposits )as tp on tp.tournamentId=t.tournamentId where t.tournamentId=?",string("T1"))
	for rows.Next() {
		var r Winner
		err = rows.Scan(&r.tournamentId, &r.playerid,&r.total,&r.backers)
		errorCheck(err)
		w = append(w, r)
	}

	return w[0]

}

type Profile struct {
	Name    int

}

func balanceHandler(w http.ResponseWriter, r *http.Request) {
	if (r.URL.Query().Get("playerId")=="") {
		http.Error(w, "No Player found", http.StatusInternalServerError);
		return
	} else {
		// need to make the int to double
		var funds int=getFundsForPlayer(r.URL.Query().Get("playerId"))
		var player string =string(r.URL.Query().Get("playerId"))
		balance:=Balance{player,funds}
		js, err := json.Marshal(balance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)

	}
}

func tourmentHandler(w http.ResponseWriter, r *http.Request) {
	//error handling
	if (r.URL.Query().Get("tournamentId") == "" || r.URL.Query().Get("playerId") == "" ) {

	} else {
		var tourmentid string="T";
		tourmentid=tourmentid+r.URL.Query().Get("tournamentId")
		var totaltourment = getTourmentRequiredMoney(tourmentid)

		if (r.URL.Query().Get("backerId") == "") {

			//goes solo
			if(playerHasEoughMoney(totaltourment, r.URL.Query().Get("playerId"))){
				 t:=Tourment{stringToInt(r.URL.Query().Get("tournamentId")),totaltourment}
				signPlayerUp(t,r.URL.Query().Get("playerId"))

			}else{
				http.Error(w, "Not enough funds", http.StatusInternalServerError)
			}

		}else{
			q := r.URL.Query()
			var tourmentBackers [] TourmentBacker
			for k := range q {

				if(k=="backerId"){
					bak:=q[k]
					var singleinput=divideMoneyperPerson(totaltourment,len(bak)+1)
					for b:=range bak{
						if(playerHasEoughMoney(singleinput,string(bak[b]))){
							tourmentBackers=AddBacker(tourmentBackers,bak[b],singleinput,true)

						}else{
							print("error")

						}
					}
					tourmentBackers=AddBacker(tourmentBackers, r.URL.Query().Get("playerId"),singleinput,false)
				}

			}
			t:=Tourment{stringToInt(r.URL.Query().Get("tournamentId")),totaltourment}
			addFundsToTourment(t,tourmentBackers,totaltourment)

		}
	}
}



func addFundsToTourment(tourment Tourment, backers []TourmentBacker,totalfunds int) {
	var player string
	var dbbackers [] TourmentBacker
	var tourmentid string="T"
	tourmentid=tourmentid+strconv.Itoa(tourment.tourmendId)
	for _, elem := range backers {
		if(!elem.isbacker){
			player=setPlayer(player,elem.Backerid)
			p:=Player{elem.Backerid,getFundsForPlayer(elem.Backerid)}
			withdrawMoney(p,(p.funds-elem.ShareofMoney))

		}else{
			p:=Player{elem.Backerid,getFundsForPlayer(elem.Backerid)}
			withdrawMoney(p,(p.funds-elem.ShareofMoney))
			dbbackers=addDBBackers(dbbackers,elem)
		}
	}

	StoreTournamentDataToDb(player,tourmentid,dbbackers,totalfunds)

}
func StoreTournamentDataToDb(playerid string,tourmentId string, dbbackers []TourmentBacker, funds int) {
	b:=DBbackers{dbbackers}
	j, err := json.Marshal(b)
	errorCheck(err)
	stmt, err := db.Prepare("insert into tourmentdeposits(tournamentId,playerid,deposit,backers) VALUES (?,?,?,?)")
	errorCheck(err)
	_, err = stmt.Exec(tourmentId, playerid,funds,string(j))
	errorCheck(err)
}

func addDBBackers(dbbackers []TourmentBacker, backer TourmentBacker) []TourmentBacker {
	return append(dbbackers, backer)
}
func setPlayer( player string, playerid string) string {
	player=playerid
	return  player
}

func AddBacker(backers []TourmentBacker, backerId string,input int,isplayer bool) []TourmentBacker{
	 return append(backers, TourmentBacker{backerId,input,isplayer})
}

func resetDBHandler(w http.ResponseWriter, r *http.Request) {
	truncateDbs()

}


func signPlayerUp(tourment Tourment, playerid string) {
	var backers[] TourmentBacker
	backers=append(backers, TourmentBacker{})
	var moneyInDb int=getFundsForPlayer(playerid);
	balance:=moneyInDb-tourment.deposit;
	p:=Player{playerid,balance}
	withdrawMoney(p,balance)
	var tourmentid string="T"
	tourmentid=tourmentid+strconv.Itoa(tourment.tourmendId)
	StoreTournamentDataToDb(playerid,tourmentid,backers,tourment.deposit)



}

func playerHasEoughMoney( totalCost int, playerId string) bool{
	var moneyInDb int=getFundsForPlayer(playerId);
	t:=moneyInDb-totalCost
	if(t>0){
		return true;
	}else{
		return false;
	}

}




func getTourmentRequiredMoney(tourmentId string) int {
	var deposit int
	err = db.QueryRow("select deposit from tourment where tournamentId=?", tourmentId).Scan(&deposit)
	switch {
	case err == sql.ErrNoRows:
		panic(err)
	default:
		return deposit
	}
}

func getFundsForPlayer(playerId string) int {
	var funds int
	err = db.QueryRow("select funds from player where playerID=?", playerId).Scan(&funds)
	switch {
	case err == sql.ErrNoRows:
		panic(err)
	default:
		return funds
	}
}





func createNewPlayer(player Player) Player {
	stmt, err := db.Prepare("insert into player(playerID,funds) VALUES (?,?)")
	errorCheck(err)
	_, err = stmt.Exec(player.playerId,0)
	errorCheck(err)
	p:=Player{player.playerId,0}
	return p;
}

func createNewTourment(tourmentId string, deposit int) {
	stmt, err := db.Prepare("insert into tourment(tournamentId,deposit) VALUES (?,?)")
	errorCheck(err)
	_, err = stmt.Exec(tourmentId, deposit)
	errorCheck(err)

}
func addFundingToPlayer(player Player, funds int) {
	stmt, err := db.Prepare("update  player set funds=? where playerID=?")
	errorCheck(err)
	_, err = stmt.Exec(funds,player.playerId)
	errorCheck(err)
}

func sameIdTourment(tourmentId string) bool {

	err = db.QueryRow("select * from tourment where tournamentId=?", tourmentId).Scan()
	switch {
	case err == sql.ErrNoRows:
		return false;
	default:
		return true
	}
}



func findUserFromDB(player Player, isfunding bool) Player {
	var dbplayer Player;
	err = db.QueryRow("select  playerid,funds from player where playerid=?", player.playerId).Scan(&dbplayer.playerId, &dbplayer.funds)
	switch {
	case err == sql.ErrNoRows:
		if (isfunding) {
			dbplayer=createNewPlayer(player)
			
		} else {
			fmt.Print(err)
		}
	case err != nil:
		fmt.Print(err)
	default:
		return dbplayer;
	}

	return dbplayer;
}
func findUserById(playerid string) Player {

	var dbplayer Player;
	err = db.QueryRow("select  playerid,funds from player where playerid=?",playerid).Scan(&dbplayer.playerId, &dbplayer.funds)
	switch {
	case err == sql.ErrNoRows:
			fmt.Print(err)

	case err != nil:
		fmt.Print(err)
	default:
		return dbplayer;
	}

	return dbplayer;

}


func withdrawMoney(player Player, funds int) {
	stmt, err := db.Prepare("update player set funds=? where playerid=?")
	errorCheck(err)
	_, err = stmt.Exec(funds, player.playerId)
	errorCheck(err)
}


func truncateDbs() {
	_, err = db.Exec("truncate table player")
	errorCheck(err)
	_, err = db.Exec("truncate table tourment")
	errorCheck(err)

	_, err = db.Exec("truncate table endedtourments")
	errorCheck(err)

	_, err = db.Exec("truncate table tourmentdeposits")
	errorCheck(err)
}



func hasEnoughMoney(player Player, dbplayer Player) bool {
	if (dbplayer.funds > player.funds) {
		return true
	}
	return false;
}


func stringToInt(val string) int{
	ival,err:=strconv.Atoi(val);
	errorCheck(err)
	return ival;
}

func divideMoneyperPerson(deposit int, numofPeople int) int {
	return deposit / numofPeople
}

func errorCheck(err error) {
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
}
