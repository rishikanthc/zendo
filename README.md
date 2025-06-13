<p align="center">
  <img src="https://raw.githubusercontent.com/rishikanthc/zendo/main/static/icon-96x96.png" width="96" style="border-radius: 50%; alt="Logo"/>
</p>
<h1 align="center">ZenDo</h1>
<p align="center">
  Minimalist weekly task manager.<br>
  <i>Plan your week, zen-style ✅</i>
</p>

ZenDo is a minimalistic task manager that is based on weekly planning. It’s a very simple and straightforward to use task manager that allows you to plan your week by assigning tasks to specific days.

ZenDo is dead simple to use. It features a very simple and beautiful UI. Simply assign tasks to days of the week. ZenDo also has PWA support, allowing you to install and use ZenDo as an app on desktop and mobile platforms.

## Screenshots

<img src="screenshots/zendo.png" alt="Zendo" width="400">
<img src="screenshots/zendo-mobile.png" alt="Zendo Mobile" width="250">

## Installation

ZenDo is distributed as a docker image and can be installed using the docker compose example shown below.
Before deploying make sure to create an empty db file using `touch local.db`. This should be a file and NOT a directory.

````yaml
version: "3.8"

services:
  app:
    image: ghcr.io/rishikanthc/zendo:v0.1.0
    ports:
      - "3000:3000"
    # (Optional) If you want to override or document it here:
    # environment:
    #   - DATABASE_URL=file:local.db
    # If you need to persist the SQLite file outside the container:
    volumes:
      - ./local.db:/app/local.db
````

## Usage

1. Add tasks by entering task description in the input area and hit enter.
1. Hovering over a task exposes controls to delete the task or move it to another day.
1. On mobile, swipe right on a task to delete and swipe left to move a task to another day.
1. ZenDo also has an 8th category called *Someday*. You can think of this as an inbox for you to add tasks temporarily and schedule them to a day later on.

## Roadmap

Below are a list of currently planned features and will be updated as the app evolves

1. Ability to add recurring tasks
1. Ability to add sub tasks
1. Set due date and dispatch reminder notifications using Ntfy, Gotify, discord etc.
1. Visualization of task statistic over time to track general efficency

# Contributing

Contributions are most welcome!
If you have any cool ideas /  any issues please open an
issue in the issue tracker and I’ll get back to you as soon as possible.
Please follow these steps to contribute to development:

1. Fork the repository.
1. Create a feature branch (git checkout -b feature/my-cool-feature).
1. Commit your changes (git commit -m “Add awesome feature”).
1. Push to your branch (git push origin feature/my-cool-feature).
1. Open a Pull Request, describing the change and any setup steps.

# License

ZenDo is licensed under the MIT license

