// import { AuthOptions, Session, getServerSession } from "next-auth";
// import GoogleProvider from "next-auth/providers/google";
// import prisma from "./db";
// import { User } from "@prisma/client";

// export const authOptions: AuthOptions = {
//   providers: [
//     GoogleProvider({
//       clientId: process.env.GOOGLE_CLIENT_ID as string,
//       clientSecret: process.env.GOOGLE_CLIENT_SECRET as string,
//     }),
//   ],
//   callbacks: {
//     async signIn({ account, profile }) {
//       if (account?.provider === "google" && profile?.email) {
//         try {
//           const user = await prisma.user.findUnique({
//             where: { email: profile.email },
//           });
//           if (!user) {
//             await prisma.user.create({
//               data: {
//                 name: profile.name,
//                 email: profile.email,
//                 image: profile.image,
//               },
//             });
//           }
//           return true;
//         } catch (error) {
//           console.error("Error during sign-in:", error);
//         }
//       }
//       return false;
//     },
//     async session({ session }) {
//       const customSession: CustomSession = session as CustomSession;
//       if (customSession.user) {
//         const user = await prisma.user.findUnique({
//           where: { email: customSession.user.email || "" },
//         });
//         customSession.user = user || undefined;
//       }
//       return customSession;
//     },

//     async jwt({ token, account }) {
//       if (account) {
//         token.sub = account.providerAccountId;
//       }
//       return token;
//     },
//   },
//   secret: process.env.NEXTAUTH_SECRET,
// };

// interface CustomSession extends Session {
//   user?: User;
// }

// export const getSession = () => getServerSession(authOptions);
