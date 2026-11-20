import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'services/api.dart';
import 'services/auth.dart';
import 'screens/login_screen.dart';
import 'screens/home_screen.dart';

void main() {
  runApp(const FlexGramApp());
}

class FlexGramApp extends StatelessWidget {
  const FlexGramApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => AuthService()),
        ProxyProvider<AuthService, ApiService>(
          update: (_, auth, __) => ApiService(auth),
        ),
      ],
      child: MaterialApp(
        title: 'FlexGram',
        debugShowCheckedModeBanner: false,
        theme: ThemeData(
          brightness: Brightness.dark,
          primaryColor: const Color(0xFF2AABEE),
          scaffoldBackgroundColor: const Color(0xFF0E1621),
          appBarTheme: const AppBarTheme(
            backgroundColor: Color(0xFF17212B),
            elevation: 0,
          ),
          colorScheme: const ColorScheme.dark(
            primary: Color(0xFF2AABEE),
            secondary: Color(0xFF2AABEE),
            surface: Color(0xFF17212B),
          ),
        ),
        home: const AuthWrapper(),
      ),
    );
  }
}

class AuthWrapper extends StatelessWidget {
  const AuthWrapper({super.key});

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthService>();
    if (auth.isLoading) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }
    return auth.isLoggedIn ? const HomeScreen() : const LoginScreen();
  }
}
