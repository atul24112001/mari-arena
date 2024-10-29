"use client";
import { User } from "@prisma/client";
import { useWallet } from "@solana/wallet-adapter-react";
import { useMutation, useQuery } from "@tanstack/react-query";
import axios, {
  AxiosHeaders,
  AxiosInstance,
  AxiosStatic,
  CreateAxiosDefaults,
  HeadersDefaults,
} from "axios";
import {
  PropsWithChildren,
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";

type AuthContextType = {
  user: User | null;
  setUser: (data: User | null) => void;
  apiClient: AxiosInstance;
};

const AuthContext = createContext<AuthContextType>({
  user: null,
  setUser: () => {},
  apiClient: axios,
});

async function verifyUser(data: {
  name: string;
  identifier: string;
}): Promise<{ data: [User]; message: string; token: string }> {
  const response = await axios.post(
    `${process.env.NEXT_PUBLIC_API_URL}/api/user`,
    data
  );
  return response.data;
}

export const AuthContextProvider = ({ children }: PropsWithChildren) => {
  const [user, setUser] = useState<null | User>(null);
  const [token, setToken] = useState<string | null>(null);

  const wallet = useWallet();

  const { mutate } = useMutation({
    mutationFn: verifyUser,
    mutationKey: ["verifyUser"],
  });

  const apiClient = useMemo(() => {
    const headers: any = {};
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    return axios.create({
      baseURL: process.env.NEXT_PUBLIC_API_URL,
      headers,
    });
  }, [token]);

  useEffect(() => {
    if (wallet.connected) {
      const publicKey = wallet.publicKey?.toBase58();
      if (publicKey) {
        mutate(
          {
            name: publicKey,
            identifier: publicKey,
          },
          {
            onError: () => {
              wallet.disconnect();
            },
            onSuccess: (data) => {
              setUser(data.data[0]);
              setToken(data.token);
            },
          }
        );
        return;
      }
      wallet.disconnect();
    } else {
      setUser(null);
    }
  }, [wallet.connected]);
  return (
    <AuthContext.Provider value={{ user, setUser, apiClient }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);
