[English version](readme-en.md)

# Скачивание видео с GetCourse без перекодирования

Данное программное обеспечение позволяет скачивать HLS видеопотоки с GetCourse без потери качества, без необходимости перекодирования и объединять их в единый видеофайл. Программа написана на языке Go и предназначена для использования в операционной системе **Windows**.

Скомпилированные бинарные файлы можно найти в [последнем релизе](https://github.com/SijyKijy/getcourse-video-downloader/releases/latest).

![](img/pic01.png)

## 0. Предварительные требования

**Для работы программы требуется `ffmpeg`.**

### – Загрузка FFmpeg

Скачайте `ffmpeg` с [https://ffmpeg.org/download.html](https://ffmpeg.org/download.html) и добавьте его в PATH системы, используя следующую команду PowerShell (запуск от администратора):

Пример команды:
```ps
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\ffmpeg\bin", "Machine")
```

## 1. Как получить ссылку на видео

* Откройте страницу с видео в браузере Chromium / Google Chrome.
* Нажмите правой кнопкой мыши на видео, выберите "Просмотр кода".
* В открывшейся панели разработчика перейдите на вкладку "Сеть" (Network).
* Перезагрузите страницу в браузере.
* Выберите желаемое разрешение видео в настройках видеоплеера GetCourse.
* Начните воспроизведение видео, дайте ему проиграть пару секунд, затем поставьте на паузу.
* Найдите и скопируйте ссылку ("Request URL") на загруженный файл с числовым названием, соответствующим разрешению видео в плеере (360, 720, 1080 и т.д.).

![](img/pic02.png)

### Другой удобный способ (Генератор)
Вы можете использовать [сайт-генератор](https://sijykijy.github.io/getcourse-video-downloader-win/website/).
Для этого нужно скопировать html-код фрейма с видео и вставить его на сайт

<img width="658" height="500" alt="image" src="https://github.com/user-attachments/assets/5dfd6163-314a-4ac3-a531-ebeed7f99c8d" />

![image](https://github.com/user-attachments/assets/80a1024a-0f4b-47c3-abd5-8b148d1ed847)

## 2. Запуск программы

### – Если вы новичок

Просто [скачайте](https://github.com/SijyKijy/getcourse-video-downloader/releases/latest) бинарный файл программы и запустите его со следующими параметрами в Windows:

`.\getcourse-video-downloader.exe 'https://player02.getcourse.ru/api/playlist/media/...' aboba.mp4`

### – Если вы разработчик

Вы знаете, что делать.

## Благодарности

Общая логика вдохновлена работой [mikhailnov](https://github.com/mikhailnov/getcourse-video-downloader), но его решение на момент публикации этого readme имеет существенную проблему со звуком в клипах после сборки (явные щелчки в начале каждого сегмента HLS видеопотока).
