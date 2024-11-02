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
  Dispatch,
  PropsWithChildren,
  SetStateAction,
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";

type AuthContextType = {
  user: User | null;
  setUser: Dispatch<SetStateAction<User | null>>;
  apiClient: AxiosInstance;
  isAdmin: boolean;
  sendMessage: (type: string, data: any) => void;
  socket: WebSocket | null;
};

const AuthContext = createContext<AuthContextType>({
  user: null,
  setUser: () => {},
  apiClient: axios,
  isAdmin: false,
  sendMessage: (type: string, data: any) => {},
  socket: null,
});

async function verifyUser(data: { name: string; identifier: string }): Promise<{
  data: [User];
  message: string;
  token: string;
  isAdmin?: boolean;
}> {
  const response = await axios.post(
    `${process.env.NEXT_PUBLIC_API_URL}/api/user`,
    data
  );
  return response.data;
}

export const AuthContextProvider = ({ children }: PropsWithChildren) => {
  const [user, setUser] = useState<null | User>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isAdmin, setIsAdmin] = useState(false);
  const [socket, setSocket] = useState<WebSocket | null>(null);

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
    if (user && !socket) {
      const ws = new WebSocket(`${process.env.NEXT_PUBLIC_API_URL}/ws`);

      ws.onopen = () => {
        setSocket(ws);
      };

      ws.onclose = () => {
        setSocket(null);
      };

      ws.onerror = () => {
        setSocket(null);
      };

      () => {
        ws.close();
      };
    }
  }, [user, socket]);

  const sendMessage = useCallback(
    (type: string, data: any) => {
      if (socket) {
        socket.send(
          JSON.stringify({
            type,
            data,
          })
        );
      }
    },
    [socket]
  );

  useEffect(() => {
    if (socket && user) {
      sendMessage("add-user", {
        userId: user.id,
        publicKey: user.email,
      });

      socket.onmessage = (e) => {
        const { type } = JSON.parse(e.data);
        if (type === "refresh") {
          window.location.reload();
        }
      };
    }
  }, [socket, user]);

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
              if (data.isAdmin) {
                setIsAdmin(true);
              }
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
    <AuthContext.Provider
      value={{ user, sendMessage, setUser, apiClient, isAdmin, socket }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);
