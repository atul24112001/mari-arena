package transaction

import (
	"errors"
	"flappy-bird-server/lib"
	"flappy-bird-server/middleware"
	"flappy-bird-server/model"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RequestBody struct {
	Signature string `json:"signature"`
}

func verifyTransaction(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.CheckAccess(w, r)
	if err != nil {
		lib.ErrorJsonWithCode(w, err, http.StatusBadRequest)
		return
	}
	var body RequestBody
	if err := lib.ReadJsonFromBody(r, w, &body); err != nil {
		lib.ErrorJsonWithCode(w, err, http.StatusBadRequest)
		return
	}

	var transactionAlreadyVerified model.Transaction
	err = lib.Pool.QueryRow(r.Context(), `SELECT id, signature, amount, "userId" FROM public.transactions WHERE signature = $1`, body.Signature).Scan(&transactionAlreadyVerified.Id, &transactionAlreadyVerified.Signature, &transactionAlreadyVerified.Amount, &transactionAlreadyVerified.UserId)
	if err != nil {
		if err == pgx.ErrNoRows {
			transaction, err := lib.GetTransaction(body.Signature)
			if err != nil {
				lib.ErrorJsonWithCode(w, err, http.StatusBadRequest)
				return
			}

			if transaction.To != lib.AdminPublicKey {
				lib.ErrorJsonWithCode(w, errors.New("Invalid transaction: please send solana to "+lib.AdminPublicKey), http.StatusBadRequest)
				return
			}

			log.Println(user.Email, transaction.From)
			if transaction.From != user.Email {
				lib.ErrorJsonWithCode(w, errors.New("Invalid transaction: user & sender details not matching"), http.StatusBadRequest)
				return
			}

			var updatedBalance int
			var amount uint
			transactionId, err := uuid.NewRandom()
			if err != nil {
				lib.ErrorJsonWithCode(w, errors.New("Something went wrong while creating transaction id"), http.StatusBadRequest)
				return
			}
			if err = lib.Pool.QueryRow(r.Context(), `INSERT INTO public.transactions (id, amount, signature, "userId") VALUES ($1, $2, $3, $4) RETURNING amount`, transactionId, transaction.Amount, body.Signature, user.Id).Scan(&amount); err != nil {
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
	}

	lib.WriteJson(w, http.StatusOK, map[string]interface{}{
		"message": "transaction verified",
		"data":    []model.User{user},
	})
}
