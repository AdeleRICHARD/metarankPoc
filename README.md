# But 

Le but de ce projet est de montrer qu'on peut utiliser metarank pour pouvoir proposer aux utilisateurs une selections d'annonces similaire par rapport à ce qu'ils cherche.

## Prérequis

Avant de commencer, assurez-vous d'avoir installé les outils suivants sur votre machine :

- [Go](https://golang.org/doc/install) (version 1.16 ou supérieure)
- [Docker](https://www.docker.com/get-started) pour exécuter Metarank
- [Metarank](https://github.com/metarank/metarank) (pour l'entraînement et le déploiement des modèles)

## How to use it

### Étape 1 : Cloner le dépôt

- Clonez le dépôt sur votre machine locale

``` bash 
git clone git@github.com:AdeleRICHARD/metarankPoc.git
cd metarankPoc
```

## Étape 2 : Configurer Metarank

### Préparer les données :

Placez vos données d'annonces dans un fichier JSONL dans le dossier `data/`.
Utilisez le script `main.go` pour créer des événements utilisateur et interaction à partir du json d'annonces. 
Il faut au moins 100 annonces différentes. 
Dans le dataset `classifieds.json` il y a 100 annonces mais avec x annonces définis sur Paris 16, x autres annonces sur Nice etc.

### Configurer Metarank :

Un fichier de config a déjà été généré dans ce projet. Mais si vous voulez changer de config en fonction de vos données, vous pouvez utiliser l'autofeature de metarank.

Exemple : 
```bash
docker run -v $(pwd)/user_interaction_events_100.jsonl:/input/events.jsonl -v $(pwd):/output metarank/metarank autofeature --data /input/events.jsonl --out /output/config.yml
```

Les données dans le jsonl doivent être triées et doivent comporter au moins 10 clicks utilisateurs et au moins 100 events. 
Il faut aussi que ça soit sorted par timestamp, mais metarank a une commande de sort.

Modifiez le fichier `config.yml` pour définir vos modèles de recommandation, tels que `similar`, `trending`, etc.

### Démarrer Metarank avec Docker :

Utilisez Docker pour lancer Metarank soit en mode serveur :

```bash
docker run -p 8080:8080 -v $(pwd)/config.yml:/config.yml -v $(pwd)/data:/data metarank/metarank serve --config /config.yml
```

Soit directement en api entrainée : 

```bash
docker run -i -t -p 8080:8080 -v $(pwd):/opt/metarank metarank/metarank:latest standalone --config /opt/metarank/config.yml --data /opt/metarank/user_interaction_events_100.jsonl.gz
```
Cette deuxième approche permet une première initialisation de métarank avec nos bonnes données et notre bonne config pour une service prêt à l'emploi. 


On pourra ensuite l'entraîner via l'appel api entraînement : 
```bash
docker run -v $(pwd)/data:/data metarank/metarank train --config /config.yml
```

## Tester le programme et l'api

Pour tester que notre modèle est entraîné correctement et qu'on a les appels api qui fonctionnent, il suffit de lancer la commande de test : 

```bash
go test -v
```

