package transaction

import (
	gameManager "flappy-bird-server/game-manager"
	"flappy-bird-server/lib"
	"flappy-bird-server/model"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
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

func verifyTransaction(c *fiber.Ctx) error {
	token := c.Get("Authorization")

	if token != os.Getenv("HELIUS_WEBHOOK_SECRET") {
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]interface{}{
			"message": "Unauthorized",
		})
	}

	var body RequestBody
	err := c.BodyParser(body)
	if err != nil {
		return c.Status(400).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}
	if len(body) == 0 {
		return c.Status(500).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}

	transaction := body[0]
	var transactionAlreadyVerified model.Transaction
	var user model.User

	err = lib.Pool.QueryRow(c.Context(), `SELECT id, signature, amount, "userId" FROM public.transactions WHERE signature = $1`, transaction.Signature).Scan(&transactionAlreadyVerified.Id, &transactionAlreadyVerified.Signature, &transactionAlreadyVerified.Amount, &transactionAlreadyVerified.UserId)
	newLine := fmt.Sprintf("ERROR_TRANSACTION-%s", transaction.Signature)
	if err != nil {
		if err == pgx.ErrNoRows {
			if len(transaction.NativeTransfers) == 0 {
				return lib.ErrorJson(c, 500, newLine+"No native transfer found\n", "transaction.txt")

			}
			transfer := transaction.NativeTransfers[0]
			newLine += fmt.Sprintf("-publicKey-%s-", transfer.FromUserAccount)
			if transfer.ToUserAccount != lib.AdminPublicKey {
				return lib.ErrorJson(c, 500, newLine+"Invalid transaction: please send solana to "+lib.AdminPublicKey+"\n", "transaction.txt")
			}

			err = lib.Pool.QueryRow(c.Context(), `SELECT id, name, email, "inrBalance", "solanaBalance" FROM public.users WHERE email = $1`, transfer.FromUserAccount).Scan(&user.Id, &user.Name, &user.Email, &user.INRBalance, &user.SolanaBalance)
			if err != nil {
				return lib.ErrorJson(c, 500, newLine+err.Error()+"\n", "transaction.txt")
			}

			var updatedBalance int
			var amount uint
			transactionId, err := uuid.NewRandom()
			if err != nil {
				return lib.ErrorJson(c, 500, newLine+"Something went wrong while creating transaction id\n", "transaction.txt")
			}
			if err = lib.Pool.QueryRow(c.Context(), `INSERT INTO public.transactions (id, amount, signature, "userId") VALUES ($1, $2, $3, $4) RETURNING amount`, transactionId, transfer.Amount, transaction.Signature, user.Id).Scan(&amount); err != nil {
				return lib.ErrorJson(c, 500, newLine+"Something went wrong while creating transaction id\n", "transaction.txt")

			}
			err = lib.Pool.QueryRow(c.Context(), `UPDATE public.users SET "solanaBalance" = $2 WHERE id = $1 RETURNING "solanaBalance"`, user.Id, user.SolanaBalance+amount).Scan(&updatedBalance)
			if err != nil {
				return lib.ErrorJson(c, 500, newLine+"Something went wrong while update user solana balance\n", "transaction.txt")

			}
			user.SolanaBalance = uint(updatedBalance)
		} else {
			return lib.ErrorJson(c, 500, newLine+"Something went wrong while fetching transaction details\n", "transaction.txt")
		}
	} else {
		err = lib.Pool.QueryRow(c.Context(), `SELECT id, name, email, "inrBalance", "solanaBalance" FROM public.users WHERE id = $1`, transactionAlreadyVerified.UserId).Scan(&user.Id, &user.Name, &user.Email, &user.INRBalance, &user.SolanaBalance)
		return lib.ErrorJson(c, 500, newLine+err.Error()+"\n", "transaction.txt")
	}

	userWs, userWsExist := gameManager.GetInstance().GetUser(user.Id)
	if userWsExist {
		userWs.SendMessage("refresh", map[string]interface{}{})
	}
	return c.Status(200).JSON(map[string]interface{}{
		"message": "transaction verified",
		"data":    []model.User{user},
	})
}
