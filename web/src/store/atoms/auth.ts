import { atom } from "recoil";
import STORE_KEYS from "../keys.json";
import { User } from "@prisma/client";

type AuthStateType = {
  isAuthenticated: boolean;
  user: User | null;
};

export const AuthState = atom<AuthStateType>({
  key: STORE_KEYS.ATOM_AUTH_STATE,
  default: {
    isAuthenticated: false,
    user: null,
  },
});
