package transaction

import (
	"errors"
	gameManager "flappy-bird-server/game-manager"
	"flappy-bird-server/lib"
	"flappy-bird-server/model"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type NativeTransfers struct {
	Amount          int    `json:"amount"`
	FromUserAccount string `json:"fromUserAccount"`
	ToUserAccount   string `json:"toUserAccount"`
}

type RequestBody []struct {
	Signature       string            `json:"signature"`
	Type            string            `json:"type"`
	NativeTransfers []NativeTransfers `json:"nativeTransfers"`
}

func verifyTransaction(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")

	if token != os.Getenv("HELIUS_WEBHOOK_SECRET") {
		log.Println("Unauth", "token", token, "envToken", os.Getenv("HELIUS_WEBHOOK_SECRET"))
		lib.ErrorJsonWithCode(w, errors.New("unauthorized"), http.StatusBadRequest)
		return
	}

	var body RequestBody
	if err := lib.ReadJsonFromBody(r, w, &body); err != nil {
		lib.ErrorJsonWithCode(w, err, http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		lib.ErrorJsonWithCode(w, errors.New("Something wen wrong, helius boy length is 0"), http.StatusBadRequest)
		return
	}

	transaction := body[0]
	var transactionAlreadyVerified model.Transaction
	log.Println(transaction.Signature)
	log.Println(transaction.Type)
	var user model.User
	err := lib.Pool.QueryRow(r.Context(), `SELECT id, signature, amount, "userId" FROM public.transactions WHERE signature = $1`, transaction.Signature).Scan(&transactionAlreadyVerified.Id, &transactionAlreadyVerified.Signature, &transactionAlreadyVerified.Amount, &transactionAlreadyVerified.UserId)
	if err != nil {
		if err == pgx.ErrNoRows {
			// _tac, err := lib.GetTransaction(transaction.Signature)
			// log.Printf("Fetched results", _tac.Amount)
			if len(transaction.NativeTransfers) == 0 {
				lib.ErrorJsonWithCode(w, errors.New("No native transfer found"), http.StatusBadRequest)
				return
			}
			transfer := transaction.NativeTransfers[0]

			if transfer.ToUserAccount != lib.AdminPublicKey {
				lib.ErrorJsonWithCode(w, errors.New("Invalid transaction: please send solana to "+lib.AdminPublicKey), http.StatusBadRequest)
				return
			}

			err = lib.Pool.QueryRow(r.Context(), `SELECT id, name, email, "inrBalance", "solanaBalance" FROM public.users WHERE email = $1`, transfer.FromUserAccount).Scan(&user.Id, &user.Name, &user.Email, &user.INRBalance, &user.SolanaBalance)
			if err != nil {
				lib.ErrorJsonWithCode(w, errors.New("internal server error"), http.StatusBadRequest)
				return
			}

			var updatedBalance int
			var amount uint
			transactionId, err := uuid.NewRandom()
			if err != nil {
				lib.ErrorJsonWithCode(w, errors.New("Something went wrong while creating transaction id"), http.StatusBadRequest)
				return
			}
			if err = lib.Pool.QueryRow(r.Context(), `INSERT INTO public.transactions (id, amount, signature, "userId") VALUES ($1, $2, $3, $4) RETURNING amount`, transactionId, transfer.Amount, transaction.Signature, user.Id).Scan(&amount); err != nil {
				lib.ErrorJsonWithCode(w, errors.New("Something went wrong while creating transaction"), http.StatusBadRequest)
				log.Println(err.Error())
				return
			}
			err = lib.Pool.QueryRow(r.Context(), `UPDATE public.users SET "solanaBalance" = $2 WHERE id = $1 RETURNING "solanaBalance"`, user.Id, user.SolanaBalance+amount).Scan(&updatedBalance)
			if err != nil {
				log.Println(err.Error())
				lib.ErrorJsonWithCode(w, errors.New("Something went wrong while update user solana balance"), http.StatusBadRequest)
				return
			}
			user.SolanaBalance = uint(updatedBalance)
		} else {
			log.Println(err.Error())
			lib.ErrorJsonWithCode(w, errors.New("Something went wrong while fetching transaction details"), http.StatusBadRequest)
			return
		}
	} else {
		err = lib.Pool.QueryRow(r.Context(), `SELECT id, name, email, "inrBalance", "solanaBalance" FROM public.users WHERE id = $1`, transactionAlreadyVerified.UserId).Scan(&user.Id, &user.Name, &user.Email, &user.INRBalance, &user.SolanaBalance)
		if err != nil {
			lib.ErrorJsonWithCode(w, errors.New("internal server error"), http.StatusBadRequest)
			return
		}
	}

	userWs, userWsExist := gameManager.GetInstance().GetUser(user.Id)
	if userWsExist {
		userWs.SendMessage("refresh", map[string]interface{}{})
	}
	lib.WriteJson(w, http.StatusOK, map[string]interface{}{
		"message": "transaction verified",
		"data":    []model.User{user},
	})
}
