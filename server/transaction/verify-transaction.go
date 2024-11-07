package transaction

import (
	gameManager "flappy-bird-server/game-manager"
	"flappy-bird-server/lib"
	"flappy-bird-server/model"
	"fmt"
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
		lib.ErrorJson(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	var body RequestBody
	err := lib.ReadJsonFromBody(r, w, &body)
	if err != nil {
		lib.ErrorJson(w, http.StatusBadRequest, err.Error(), "")
		return
	}

	if len(body) == 0 {
		lib.ErrorJson(w, http.StatusBadRequest, err.Error(), "")
		return
	}

	transaction := body[0]
	var transactionAlreadyVerified model.Transaction
	var user model.User

	err = lib.Pool.QueryRow(r.Context(), `SELECT id, signature, amount, "userId" FROM public.transactions WHERE signature = $1`, transaction.Signature).Scan(&transactionAlreadyVerified.Id, &transactionAlreadyVerified.Signature, &transactionAlreadyVerified.Amount, &transactionAlreadyVerified.UserId)
	newLine := fmt.Sprintf("ERROR_TRANSACTION-%s", transaction.Signature)
	if err != nil {
		if err == pgx.ErrNoRows {
			if len(transaction.NativeTransfers) == 0 {
				lib.ErrorJson(w, 500, newLine+"No native transfer found\n", "transaction.txt")
				return

			}
			transfer := transaction.NativeTransfers[0]
			newLine += fmt.Sprintf("-publicKey-%s-", transfer.FromUserAccount)
			if transfer.ToUserAccount != lib.AdminPublicKey {
				lib.ErrorJson(w, 500, newLine+"Invalid transaction: please send solana to "+lib.AdminPublicKey+"\n", "transaction.txt")
				return
			}

			err = lib.Pool.QueryRow(r.Context(), `SELECT id, name, email, "inrBalance", "solanaBalance" FROM public.users WHERE email = $1`, transfer.FromUserAccount).Scan(&user.Id, &user.Name, &user.Email, &user.INRBalance, &user.SolanaBalance)
			if err != nil {
				lib.ErrorJson(w, 500, newLine+err.Error()+"\n", "transaction.txt")
				return
			}

			var updatedBalance int
			var amount uint
			transactionId, err := uuid.NewRandom()
			if err != nil {
				lib.ErrorJson(w, 500, newLine+"Something went wrong while creating transaction id\n", "transaction.txt")
				return
			}
			if err = lib.Pool.QueryRow(r.Context(), `INSERT INTO public.transactions (id, amount, signature, "userId") VALUES ($1, $2, $3, $4) RETURNING amount`, transactionId, transfer.Amount, transaction.Signature, user.Id).Scan(&amount); err != nil {
				lib.ErrorJson(w, 500, newLine+"Something went wrong while creating transaction id\n", "transaction.txt")
				return

			}
			err = lib.Pool.QueryRow(r.Context(), `UPDATE public.users SET "solanaBalance" = $2 WHERE id = $1 RETURNING "solanaBalance"`, user.Id, user.SolanaBalance+amount).Scan(&updatedBalance)
			if err != nil {
				lib.ErrorJson(w, 500, newLine+"Something went wrong while update user solana balance\n", "transaction.txt")
				return

			}
			user.SolanaBalance = uint(updatedBalance)
		} else {
			lib.ErrorJson(w, 500, newLine+"Something went wrong while fetching transaction details\n", "transaction.txt")
			return
		}
	} else {
		err = lib.Pool.QueryRow(r.Context(), `SELECT id, name, email, "inrBalance", "solanaBalance" FROM public.users WHERE id = $1`, transactionAlreadyVerified.UserId).Scan(&user.Id, &user.Name, &user.Email, &user.INRBalance, &user.SolanaBalance)
		lib.ErrorJson(w, 500, newLine+err.Error()+"\n", "transaction.txt")
		return
	}

	userWs, userWsExist := gameManager.GetInstance().GetUser(user.Id)
	if userWsExist {
		userWs.SendMessage("refresh", map[string]interface{}{})
	}
	lib.WriteJson(w, 200, map[string]interface{}{
		"message":   "transaction verified",
		"data":      []model.User{user},
		"signature": transaction.Signature,
	})
}
