import { selector, useRecoilState, useRecoilValue } from "recoil";
import { AuthState } from "..";
import STORE_KEYS from "../keys.json";
import { User } from "@prisma/client";

const isAuthenticatedState = selector({
  key: STORE_KEYS.SELECTOR_AUTH_IS_AUTHENTICATED_STATE,
  get: ({ get }) => {
    return get(AuthState).isAuthenticated;
  },
  set: ({ get, set }, value) => {
    set(AuthState, {
      ...get(AuthState),
      isAuthenticated: value as boolean,
    });
  },
});

const userState = selector({
  key: STORE_KEYS.SELECTOR_AUTH_USER_STATE,
  get: ({ get }) => {
    return get(AuthState).user;
  },
  set: ({ get, set }, value) => {
    set(AuthState, {
      ...get(AuthState),
      user: value as User | null,
    });
  },
});

// export const useIsAuthenticatedState = () =>
//   useRecoilState(isAuthenticatedState);
// export const useIsAuthenticatedValue = () =>
//   useRecoilValue(isAuthenticatedState);
// export const useSetIsAuthenticated = () => useRecoilValue(isAuthenticatedState);

// export const useUserState = () => useRecoilState(userState);
// export const useUserValue = () => useRecoilValue(userState);
// export const useSetUser = () => useRecoilValue(userState);
