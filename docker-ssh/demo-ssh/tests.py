#!/usr/bin/env python3

import unittest
import requests
import warnings
from progressbar import ProgressBar, Percentage, Bar, ETA
from colorama import init, Fore
import time

from tqdm import tqdm


OK_STATUS_CODE = 200
BAD_REQUEST_STATUS_CODE = 400
UNAUTHORIZED_STATUS_CODE = 401
NOT_FOUND_STATUS_CODE = 404
CONFLICT_STATUS_CODE = 409
INTERNAL_SERVER_ERROR_STATUS_CODE = 500

# Inicializar colorama para la impresión de colores en la consola
init(autoreset=True)

# Deshabilitar advertencias de SSL
warnings.filterwarnings("ignore")


class APITests(unittest.TestCase):
    access_token = None

    def setUp(self):
        # Configuración inicial, se ejecuta antes de cada prueba
        self.base_url = "https://myserver.local:5000"
        # Deshabilitar la verificación del certificado SSL en entornos de prueba
        self.verify_ssl = False

    def tearDown(self):
        # Limpieza después de cada prueba
        pass

    def test_get_version(self):
        # Prueba para la ruta /version
        url = f"{self.base_url}/version"
        response = requests.get(url, verify=self.verify_ssl)

        self.assertEqual(response.status_code, OK_STATUS_CODE)
        # Permitir varias respuestas correctas

        self.assertIn("version", response.json())
        time.sleep(2)

    def test_sign_up_new_user(self):
        # Prueba para la ruta /signup
        url = f"{self.base_url}/signup"
        data = {"username": "test_user", "password": "test_password"}
        response = requests.post(url, json=data, verify=self.verify_ssl)

        expected_status_codes = [OK_STATUS_CODE, CONFLICT_STATUS_CODE]

        self.assertIn(response.status_code, expected_status_codes)
        # Permitir varias respuestas correctas
        if response.status_code == OK_STATUS_CODE:
            self.assertIn("access_token", response.json())
        if response.status_code == CONFLICT_STATUS_CODE:
            print("[SIGNUP] User already exists.")
            self.assertEqual(response.json()["error"], "User already exists.")

        time.sleep(2)

    def test_login_user(self):
        url = f"{self.base_url}/login"
        data = {"username": "test_user", "password": "test_password"}
        response = requests.post(url, json=data, verify=self.verify_ssl)
        self.access_token = response.json()["access_token"]
        print("access_token: ", self.access_token)
        self.assertEqual(response.status_code, OK_STATUS_CODE)
        # Permitir varias respuestas correctas
        self.assertIn("access_token", response.json())

        time.sleep(2)

    def test_post_file(self):
        # Prueba para la ruta /test_user/test_doc
        header = introduce_access_token(self.access_token)
        url = f"{self.base_url}/test_user/test_doc_new"
        data = {"doc_content": "test_content"}

        expected_status_codes = [OK_STATUS_CODE, BAD_REQUEST_STATUS_CODE]
        response = requests.post(
            url, json=data, headers=header, verify=self.verify_ssl)
        self.assertIn(response.status_code, expected_status_codes)
        if response.status_code == OK_STATUS_CODE:
            self.assertIn("size", response.json())
        if response.status_code == BAD_REQUEST_STATUS_CODE:
            print("[POST_FILE] Bad request.")
            self.assertIn(response.json()["error"], "Doc already exists.")

    def test_get_doc(self):

        header = introduce_access_token(self.access_token)
        url = f"{self.base_url}/test_user/test_doc"
        
        expected_status_codes = [OK_STATUS_CODE, NOT_FOUND_STATUS_CODE]
        
        response = requests.get(url, headers=header, verify=self.verify_ssl)
        print(response.json())
        if response.status_code == OK_STATUS_CODE:
            self.assertIn("doc_content", response.json())
        if response.status_code == NOT_FOUND_STATUS_CODE:
            print("[GET_DOC] Doc not found.")
            self.assertIn(response.json()["error"], "Wrong id for this user, doc not found or removed.")

    def test_get_all_docs(self):
        header = introduce_access_token(self.access_token)

        url = f"{self.base_url}/test_user/_all_docs"
        response = requests.get(url, headers=header, verify=self.verify_ssl)
        print(response.json())
        self.assertEqual(response.status_code, OK_STATUS_CODE)

    def test_delete_doc(self):
        header = introduce_access_token(self.access_token)
        url = f"{self.base_url}/test_user/test_doc_new"
        response = requests.delete(url, headers=header, verify=self.verify_ssl)
        print(response.json())
        self.assertEqual(response.status_code, OK_STATUS_CODE)
        # quiero que compruebe si hay exactamente una llave vacía: {}
        self.assertEqual(response.json(), {})


def introduce_access_token(access_token):
    access_token = input("Introduce access_token: ")
    access_token = "token " + access_token
    header = {"Authorization": access_token}
    return header

    # Agrega más pruebas según sea necesario


if __name__ == '__main__':
    # Crear una suite de pruebas y agregarlas en el orden deseado
    suite = unittest.TestSuite()
    suite.addTest(APITests("test_get_version"))
    suite.addTest(APITests("test_sign_up_new_user"))
    suite.addTest(APITests("test_login_user"))
    suite.addTest(APITests("test_post_file"))
    suite.addTest(APITests("test_get_doc"))
    suite.addTest(APITests("test_get_all_docs"))
    suite.addTest(APITests("test_delete_doc"))

    result = unittest.TestResult()

    pbar = tqdm(total=len(suite._tests), desc="Pruebas",
                bar_format="{desc}: {percentage:3.0f}%|{bar}| {n_fmt}/{total_fmt}")

    # Crear un objeto TestResult para rastrear los resultados de las pruebas
    for test in suite:
        # En lugar de addError y addFailure, simplemente ejecutamos la prueba
        test(result)
        pbar.update(1)

    # Imprimir el número de pruebas pasadas y el número total de pruebas
    passed_tests = result.testsRun - len(result.errors) - len(result.failures)
    total_tests = result.testsRun
    print(
        f"\n\n{Fore.GREEN}Tests pasados/tests totales: {passed_tests}/{total_tests}{Fore.RESET}")

    # Imprimir detalles de pruebas que fallaron
    if result.errors or result.failures:
        print(f"\n{Fore.RED}Pruebas que fallaron:{Fore.RESET}")
        for failure in result.errors:
            print(f"\n{Fore.RED}Error en prueba: {failure[0]}{Fore.RESET}")
            print(f"Detalle del error: {failure[1]}")
        for failure in result.failures:
            print(f"\n{Fore.RED}Fallo en prueba: {failure[0]}{Fore.RESET}")
            print(f"Detalle del fallo: {failure[1]}")
